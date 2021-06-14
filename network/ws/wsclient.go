package ws

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/nomos/go-events"
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/network"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/promise"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"time"
)

const (
	TimeOut = time.Second * 3
)

var _ lokas.INetClient = (*WsClient)(nil)

type WsClient struct {
	events.EventEmmiter
	session lokas.ISession
	conn           *websocket.Conn
	ws             *WebSocket
	timeout        time.Duration
	addr           string
	idGen          uint32
	context lokas.IReqContext
	reqContexts    map[uint32]lokas.IReqContext
	isOpen         bool
	Closing        bool
	Opening        bool
	done           chan struct{}
	contextMutex   sync.Mutex
	openingPending *promise.Promise
	closePending   *promise.Promise
}

func NewWsClient() *WsClient {
	ret := &WsClient{
		EventEmmiter: events.New(),
		context: network.NewDefaultContext(context.TODO()),
		reqContexts:  make(map[uint32]lokas.IReqContext),
		timeout:      TimeOut,
		isOpen:       false,
		session:network.NewDefaultSession(nil,0,nil),
	}
	ret.init()
	return ret
}

func (this *WsClient) init() {

}

func (this *WsClient) genId() uint32 {
	this.idGen++
	return this.idGen
}

func (this *WsClient) Connected() bool {
	return this.isOpen
}

func (this *WsClient) Connect(addr string) *promise.Promise {
	addr = "ws://" + addr + "/ws"
	if this.addr != "" && this.addr != addr {
		return this.Close().Catch(func(err error) interface{} {
			log.Error(err.Error())
			return nil
		}).Then(func(data interface{}) interface{} {
			this.addr = addr
			return this.Open()
		})
	}
	this.addr = addr
	return this.Open()
}

func (this *WsClient) ClearContext() {
	this.contextMutex.Lock()
	defer this.contextMutex.Unlock()
	for _, v := range this.reqContexts {
		v.Cancel(errors.New("clear context"))
	}
}

func (this *WsClient) Disconnect(force bool) *promise.Promise {
	if this.isOpen {
		this.isOpen = false

		this.Closing = true
		return this.Close()
	}
	if force {
		if this.openingPending != nil {
			this.openingPending.Reject("force disconnect")
		}
	}
	return this.Close()
}

func (this *WsClient) Run() {
	this.done = make(chan struct{}, 0)
	go func() {
		for {
			select {

			case <-this.done:
				return
			}
		}
	}()
}

func (this *WsClient) onerror() {

}

func (this *WsClient) Open() *promise.Promise {
	if this.isOpen {
		return promise.Resolve(nil)
	}
	if this.openingPending == nil {
		this.openingPending = promise.Async(func(resolve func(interface{}), reject func(interface{})) {
			timeout := promise.SetTimeout(TimeOut, func() {
				reject("timeout")
				this.openingPending = nil
			})
			this.Opening = true
			ws, err := NewWebSocket(this.addr, this)
			if err != nil {
				log.Error("create ws error", zap.String("err", err.Error()))
				reject(err)
				this.openingPending = nil
				return
			}
			this.ws = ws

			this.Once("open", func(i ...interface{}) {
				timeout.Close()
				this.openingPending = nil
				resolve(nil)
			})
		})

	}
	return this.openingPending
}

func (this *WsClient) Close() *promise.Promise {
	this.isOpen = false
	if this.closePending == nil {
		this.closePending = promise.Async(func(resolve func(interface{}), reject func(interface{})) {
			if this.ws == nil {
				resolve(nil)
				this.closePending = nil
				return
			}
			_, err := this.ws.Close().Await()
			if err != nil {
				reject(err)
				this.openingPending = nil
				return
			}
			this.Closing = false
			resolve(nil)
			this.closePending = nil
		})
	}
	return this.closePending
}

func (this *WsClient) OnOpen(conn *websocket.Conn) {
	this.conn = this.ws.Conn
	this.Opening = false
	log.Warnf(this.addr + " connected")
	this.isOpen = true
	this.Emit("open")
}

func (this *WsClient) OnClose(conn *websocket.Conn) {
	if this.isOpen {
		this.Disconnect(true)
	}
	this.conn = nil
	this.ws = nil
	this.Closing = false
	log.Warnf(this.addr + " disconnected")
	this.Emit("close")
}

func (this *WsClient) OnMessage(cmdId protocol.BINARY_TAG, listeners ...events.Listener) {
	this.On(events.EventName("CmdId"+strconv.Itoa(int(cmdId))), listeners...)
}

func (this *WsClient) Off(cmdId uint16, listener events.Listener) {
	if listener != nil {
		this.RemoveListener(events.EventName("CmdId"+strconv.Itoa(int(cmdId))), listener)
	} else {
		this.RemoveAllListeners(events.EventName("CmdId" + strconv.Itoa(int(cmdId))))
	}
}

func (this *WsClient) SendMessage(transId uint32, msg interface{}) {
	rb, err := protocol.MarshalBinaryMessage(transId, msg)
	if err != nil {
		log.Error(err.Error())
		return
	}
	this.conn.WriteMessage(websocket.BinaryMessage, rb)
}

func (this *WsClient) addContext(transId uint32, ctx lokas.IReqContext) {
	this.contextMutex.Lock()
	defer this.contextMutex.Unlock()
	this.reqContexts[transId] = ctx
}

func (this *WsClient) removeContext(transId uint32) {
	this.contextMutex.Lock()
	defer this.contextMutex.Unlock()
	delete(this.reqContexts, transId)
}

func (this *WsClient) getContext(transId uint32) lokas.IReqContext {
	this.contextMutex.Lock()
	defer this.contextMutex.Unlock()
	return this.reqContexts[transId]
}

func (this *WsClient) Request(req interface{}) *promise.Promise {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		if this.Opening {
			_, err := this.Open().Await()
			if err != nil {
				reject(err)
				return
			}
		} else if !this.Connected() {
			//log.Warn("connection closed",this)
			reject(errors.New("connection closed"))
			return
		}
		id := this.genId()
		resp,err := this.Call(id, req)
		if err != nil {
			log.Error("Call Error:%s", zap.String("err", err.Error()))
			reject(err)
			return
		}
		resolve(resp)
	})
}

func (this *WsClient) Call(transId uint32, req interface{}) (interface{},error) {
	ctx := network.NewDefaultContextWithTimeout(this.context,transId,this.timeout)
	return this.doCall(ctx, req,true)
}

func (this *WsClient) AsyncCall(transId uint32, req interface{}) (interface{},error) {
	ctx := network.NewDefaultContextWithTimeout(this.context,transId,this.timeout)
	return this.doCall(ctx, req,false)
}

func (this *WsClient) doCall(ctx lokas.IReqContext, req interface{}, isSync bool) (interface{},error) {
	transId := ctx.GetTransId()
	this.addContext(transId, ctx)
	//cmdId, err := protocol.GetCmdIdFromType(req)
	//if err != nil {
	//	log.Error(err.Error())
	//	return err
	//}
	rb, err := protocol.MarshalBinaryMessage(transId, req)
	if err != nil {
		log.Error(err.Error())
		return nil,err
	}
	this.ws.writeChan <- rb
	if !isSync {
		return nil,nil
	}
	select {
	case <-ctx.Done():
		switch ctx.Err() {
		case context.DeadlineExceeded:
			this.removeContext(transId)
			if this.isOpen {
				this.Close().Await()
				go func() {
					this.Open().Await()
				}()
			}
			return nil,protocol.ERR_RPC_TIMEOUT
		default:
			resp:=ctx.GetResp()
			if err,ok:=resp.(*protocol.ErrMsg);ok {
				return nil,err
			}
			return ctx.GetResp(),nil
		}
	}
}


func (this *WsClient) OnRecv(connect lokas.IConn,data []byte) {
	_, err := this.HookRecv(data)
	if err != nil {
		log.Error("WebSocket OnRecv err:" + err.Error())
	}
}

func (this *WsClient) OnRecvData(data []byte) {
	_, err := this.HookRecv(data)
	if err != nil {
		log.Error("WebSocket OnRecv err:" + err.Error())
	}
}

func (this *WsClient) OnRecvMessage(cmdId protocol.BINARY_TAG, transId uint32, msg interface{}) {
	this.Emit(events.EventName("CmdId"+strconv.Itoa(int(cmdId))), msg, transId)
}

func (this *WsClient) OnRecvCmd(cmdId protocol.BINARY_TAG, time time.Duration) *promise.Promise {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		timeout := promise.SetTimeout(time, func() {
			reject("timeout")
		})
		this.Once(events.EventName("CmdId"+strconv.Itoa(int(cmdId))), func(i ...interface{}) {
			timeout.Close()
			msg := i[0]
			resolve(msg)
			return
		})
	})
}

func (this *WsClient) HookRecv(data []byte) (interface{}, error) {
	msg, err := protocol.UnmarshalBinaryMessage(data)
	if err != nil {
		return nil, err
	}
	if msg.TransId == 0 {
		this.OnRecvMessage(msg.CmdId, msg.TransId, msg.Body)
		return msg, nil
	}
	if msg.TransId != 0 {
		ctx := this.getContext(msg.TransId)
		if ctx == nil {
			log.Error("msgCmdId:%d TransId :%d ctx not found",
				zap.Any("cmdId", msg.CmdId),
				zap.Uint32("transId", msg.TransId),
			)
			return msg, errors.New(fmt.Sprintf("msgCmdId:%d TransId :%d ctx not found", msg.CmdId, msg.TransId))
		}
		if msg.CmdId == protocol.TAG_Error {
			log.Error("CmdErrorAckId")
			body := msg.Body.(*protocol.ErrMsg)
			ctx.Cancel(&protocol.ErrMsg{
				Code:    body.Code,
				Message: "code:" + strconv.Itoa(int(body.Code)) + "," + body.Message,
			})
			return nil, err
		}
		this.removeContext(ctx.GetTransId())
		ctx.SetResp(msg.Body)
		ctx.Finish()
		this.OnRecvMessage(msg.CmdId, msg.TransId, msg.Body)
		return msg, nil
	}
	return nil, nil
}
