package lox

import (
	"context"
	"errors"
	"github.com/nomos/go-events"
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/network"
	"github.com/nomos/go-lokas/network/conn"
	"github.com/nomos/go-lokas/network/tcp"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/promise"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"time"
)

const (
	PingInterval = time.Second * 3
)


var _ lokas.INetClient = (*TcpClient)(nil)
var _ lokas.ISession = (*TcpClient)(nil)


type TcpClient struct {
	events.EventEmmiter
	*ActiveSession
	conn        *conn.Conn
	timeout     time.Duration
	addr        string
	idGen       uint32
	isOpen         bool
	Closing        bool
	Opening        bool
	Protocol    protocol.TYPE
	context     lokas.IReqContext
	reqContexts map[uint32]lokas.IReqContext
	openPending *promise.Promise
	closePending *promise.Promise
	mu sync.Mutex
	ctxMutex sync.Mutex
}

func tcpSessionCreator(client *TcpClient) func(conn lokas.IConn) lokas.ISession {
	return func(connect lokas.IConn) lokas.ISession {
		client.conn = connect.(*conn.Conn)
		client.ActiveSession.Conn = connect
		return client
	}
}

func NewTcpClient() *TcpClient {
	ret := &TcpClient{
		EventEmmiter:  events.New(),
		ActiveSession: NewActiveSession(nil,0,nil),
		context:       nil,
		reqContexts:   make(map[uint32]lokas.IReqContext),
		timeout:       TimeOut,
	}
	return ret
}

func (this *TcpClient) genId() uint32 {
	this.idGen++
	return this.idGen
}

func (this *TcpClient) connect()error {
	context := &lokas.Context{
		SessionCreator: tcpSessionCreator(this),
		Splitter:       protocol.Split,
		ReadBufferSize: 1024 * 1024 * 4,
		ChanSize:       10000,
		LongPacketPicker:  protocol.PickLongPacket(this.Protocol),
		LongPacketCreator: protocol.CreateLongPacket(this.Protocol),
		MaxPacketWriteLen: protocol.DEFAULT_PACKET_LEN,
	}
	connect,err:=tcp.Dial(this.addr,context)
	if err!=nil {
		return err
	}
	this.conn = connect
	this.Start()
	return nil
}

func (this *TcpClient) Connect(addr string) *promise.Promise {
	this.addr = addr
	this.mu.Lock()
	defer this.mu.Unlock()
	if this.context == nil {
		this.context = network.NewDefaultContext(context.TODO())
		if this.openPending!=nil {
			panic("err openPending should be nil")
		}
		this.openPending = promise.Async(func(resolve func(interface{}), reject func(interface{})) {
			go func() {
				err:=this.connect()
				if err != nil {
					log.Error(err.Error())
					reject(err)
					return
				}
				for {
					if this.isOpen {
						resolve(nil)
						return
					}
					time.Sleep(1)
				}
			}()
		})
		return this.openPending
	}
	if this.openPending!=nil {
		return this.openPending
	} else {
		return promise.Resolve(nil)
	}
}

func (this *TcpClient) ClearContext(err error) {
	this.ctxMutex.Lock()
	defer this.ctxMutex.Unlock()
	for _, v := range this.reqContexts {
		v.Cancel(err)
	}
	this.reqContexts = map[uint32]lokas.IReqContext{}
	this.context = nil
}

func (this *TcpClient) Disconnect(b bool) *promise.Promise {
	if this.closePending!=nil {
		return this.closePending
	}
	this.closePending = promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		this.conn.Close()
		resolve(nil)
	})
	return this.closePending
}

func (this *TcpClient) Connected() bool {
	return this.isOpen
}

func (this *TcpClient) OnOpen(conn lokas.IConn) {
	log.Warnf("TcpClient OnOpen")
	this.MsgHandler = this.MessageHandler
	this.Opening = false
	log.Warnf(this.addr + " connected")
	this.isOpen = true
	this.Emit("open")
}

func (this *TcpClient) OnClose(conn lokas.IConn) {
	log.Warnf("TcpClient OnClose")
	this.openPending = nil
	this.Opening = false
	this.ActiveSession.stop()
	this.ClearContext(errors.New("disconnect"))
	if this.isOpen {
		this.isOpen = false
		this.Disconnect(true)
		this.closePending = nil
	}
	this.conn.Close()
	this.conn = nil
	this.Closing = false
	log.Warnf(this.addr + " disconnected")
	this.Emit("close")
}

func (this *TcpClient) Request(req interface{}) *promise.Promise {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		if this.Opening {
			_, err := this.Connect(this.addr).Await()
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
		var err error
		resp,err := this.Call(id, req)
		if err != nil {
			log.Error("Call Error:%s", zap.String("err", err.Error()))
			reject(err)
			return
		}
		resolve(resp)
	})
}

func (this *TcpClient) OnRecvCmd(cmdId protocol.BINARY_TAG, time time.Duration) *promise.Promise {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		timeout := promise.SetTimeout(time, func() {
			reject(context.DeadlineExceeded)
		})
		this.Once(events.EventName("CmdId"+strconv.Itoa(int(cmdId))), func(i ...interface{}) {
			timeout.Close()
			msg := i[0]
			resolve(msg)
			return
		})
	})
}

func (this *TcpClient) MessageHandler(msg *protocol.BinaryMessage){
	id,_:=msg.GetId()
	log.Warnf("MessageHandler",msg.TransId,id)
	if msg.TransId!=0 {
		ctx:=this.GetContext(msg.TransId)
		ctx.SetResp(msg.Body)
		ctx.Finish()
	}
}

func (this *TcpClient) OnMessage(cmdId protocol.BINARY_TAG, listeners ...events.Listener) {
	this.On(events.EventName("CmdId"+strconv.Itoa(int(cmdId))),listeners...)
}

func (this *TcpClient) Off(cmdId uint16, listener events.Listener) {
	if listener != nil {
		this.RemoveListener(events.EventName("CmdId"+strconv.Itoa(int(cmdId))), listener)
	} else {
		this.RemoveAllListeners(events.EventName("CmdId" + strconv.Itoa(int(cmdId))))
	}
}

func (this *TcpClient) SendMessage(transId uint32, msg interface{}) {
	if this.Protocol == protocol.JSON {
		this.sendJsonMessage(transId,msg)
	} else if this.Protocol == protocol.BINARY {
		this.sendBinaryMessage(transId,msg)
	} else {
		panic("unidentified protocol")
	}
}

func (this *TcpClient) sendJsonMessage(transId uint32, msg interface{}) {
	rb, err := protocol.MarshalJsonMessage(transId, msg)
	if err != nil {
		log.Error(err.Error())
		return
	}
	this.conn.Write(rb)
}

func (this *TcpClient) sendBinaryMessage(transId uint32, msg interface{}) {
	rb, err := protocol.MarshalBinaryMessage(transId, msg)
	if err != nil {
		log.Error(err.Error())
		return
	}
	this.conn.Write(rb)
}

func (this *TcpClient) addContext(transId uint32, ctx lokas.IReqContext) {
	this.ctxMutex.Lock()
	defer this.ctxMutex.Unlock()
	this.reqContexts[transId] = ctx
}

func (this *TcpClient) removeContext(transId uint32) {
	this.ctxMutex.Lock()
	defer this.ctxMutex.Unlock()
	delete(this.reqContexts, transId)
}

func (this *TcpClient) GetContext(transId uint32) lokas.IReqContext {
	this.ctxMutex.Lock()
	defer this.ctxMutex.Unlock()
	return this.reqContexts[transId]
}

func (this *TcpClient) Call(transId uint32, req interface{}) (interface{},error) {
	ctx := network.NewDefaultContextWithTimeout(context.TODO(), transId, this.timeout)
	return this.doCall(ctx, req)
}

func (this *TcpClient) doCall(ctx lokas.IReqContext, req interface{}) (interface{},error) {
	transId := ctx.GetTransId()
	this.addContext(transId, ctx)
	//cmdId, err := protocol.GetCmdIdFromType(req)
	//if err != nil {
	//	log.Error(err.Error())
	//	return err
	//}

	rb, err := protocol.MarshalMessage(transId, req,this.Protocol)
	if err != nil {
		log.Error(err.Error())
		return nil,err
	}
	_, err = this.conn.Write(rb)
	if err != nil {
		log.Error(err.Error())
		return nil,err
	}
	select {
	case <-ctx.Done():
		switch ctx.Err() {
		case context.DeadlineExceeded:
			log.Warnf("DeadlineExceeded")
			this.removeContext(transId)
			if this.isOpen {
				this.Disconnect(false).Await()
				go func() {
					this.Connect(this.addr).Await()
				}()
			}
			return nil,protocol.ERR_RPC_TIMEOUT
		default:
			this.removeContext(transId)
			if ctx.Err()!=nil {
				return nil,ctx.Err()
			}
			resp:=ctx.GetResp()
			if err,ok:=resp.(*protocol.ErrMsg);ok {
				return nil,err
			}
			return ctx.GetResp(),nil
		}
	}

}
