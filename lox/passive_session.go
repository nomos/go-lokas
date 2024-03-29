package lox

import (
	"encoding/json"
	"github.com/nomos/go-lokas/log/flog"
	"time"

	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
)

type PassiveSessionOption func(*PassiveSession)

var _ lokas.ISession = &PassiveSession{}
var _ lokas.IActor = &PassiveSession{}

func NewPassiveSession(conn lokas.IConn, id util.ID, manager lokas.ISessionManager, opts ...PassiveSessionOption) *PassiveSession {
	s := &PassiveSession{
		Actor:    NewActor(),
		Messages: make(chan []byte, 100),
		Conn:     conn,
		Manager:  manager,
		timeout:  TimeOut,
		ticker:   time.NewTicker(UpdateTime),
	}
	for _, o := range opts {
		o(s)
	}
	s.SetType("PassiveSession")
	s.SetId(id)
	return s
}

type PassiveSession struct {
	*Actor
	Verified         bool
	Messages         chan []byte
	Conn             lokas.IConn
	Protocol         protocol.TYPE
	Manager          lokas.ISessionManager
	doneClient       chan struct{}
	doneServer       chan struct{}
	OnCloseFunc      func(conn lokas.IConn)
	OnOpenFunc       func(conn lokas.IConn)
	ClientMsgHandler func(msg *protocol.BinaryMessage)
	AuthFunc         func(data []byte) (interface{}, error)
	timeout          time.Duration
	ticker           *time.Ticker
}

func (this *PassiveSession) Load(conf lokas.IConfig) error {
	panic("implement me")
}

func (this *PassiveSession) Unload() error {
	panic("implement me")
}

func (this *PassiveSession) OnStart() error {
	panic("implement me")
}

func (this *PassiveSession) OnStop() error {
	panic("implement me")
}

func (this *PassiveSession) OnCreate() error {
	return nil
}

func (this *PassiveSession) Start() error {
	return nil
}

func (this *PassiveSession) Stop() error {
	return nil
}

func (this *PassiveSession) OnDestroy() error {
	return nil
}

func (this *PassiveSession) GetConn() lokas.IConn {
	return this.Conn
}

func (this *PassiveSession) StartMessagePump() {
	log.Info("PassiveSession:StartMessagePump", lokas.LogActorInfo(this)...)

	this.MsgChan = make(chan *protocol.RouteMessage, 100)
	this.doneClient = make(chan struct{})
	this.doneServer = make(chan struct{})

	go func() {
		defer func() {
			r := recover()
			if r != nil {
				if e, ok := r.(error); ok {
					log.Errorf(e.Error())
					log.Error("客户端协议出错")
					this.Conn.Close()
				}
			}
		}()
		this.clientLoop()
		close(this.doneClient)
		this.doneClient = nil
		close(this.Messages)
		this.Messages = nil

	}()

	go func() {
		defer func() {
			r := recover()
			if r != nil {
				if util.Recover(r, true) != nil {
					log.Error("服务端协议出错")
					this.Conn.Close()
				}
			}
		}()
		this.inLoop()
		close(this.MsgChan)
		this.MsgChan = nil
		close(this.doneServer)
		this.doneServer = nil
	}()
}

func (this *PassiveSession) clientLoop() {
	for {
		select {
		case <-this.ticker.C:
			if this.OnUpdateFunc != nil && this.Verified {
				this.OnUpdateFunc()
			}
		case data := <-this.Messages:
			cmdId := protocol.GetCmdId16(data)
			if !this.Verified && cmdId != protocol.TAG_HandShake {
				var msg []byte
				msg, err := protocol.MarshalMessage(0, protocol.NewError(protocol.ERR_AUTH_FAILED), this.Protocol)
				if err != nil {
					log.Error(err.Error())
					this.Conn.Wait()
					this.Conn.Close()
					return
				}
				log.Error("Auth Failed", lokas.LogActorInfo(this).Append(protocol.LogCmdId(cmdId))...)
				this.Conn.Write(msg)
				this.Conn.Wait()
				this.Conn.Close()
				return
			}
			msg, err := protocol.UnmarshalMessage(data, this.Protocol)
			if err != nil {
				log.Error("unmarshal client message error",
					lokas.LogActorInfo(this).Append(protocol.LogCmdId(cmdId))...,
				)
				msg1, _ := protocol.NewError(protocol.ERR_MSG_FORMAT).Marshal()
				_, err = this.Conn.Write(msg1)
				if err != nil {
					log.Error(err.Error())
				}
				this.Conn.Close()
				return
			}
			if cmdId == protocol.TAG_HandShake {
				if this.Verified {
					log.Warn("duplicated handshake")
					return
				}

				var ret interface{}
				if this.AuthFunc != nil {
					ret, err = this.AuthFunc(msg.Body.(*protocol.HandShake).Data)
				}
				if err != nil {
					log.Error(err.Error())
					msg1, err1 := protocol.MarshalMessage(msg.TransId, protocol.NewError(protocol.ERR_AUTH_FAILED), this.Protocol)
					if err1 != nil {
						log.Error(err1.Error())
						this.Conn.Wait()
						this.Conn.Close()
						return
					}
					log.Warn("Auth Failed", lokas.LogActorInfo(this).Append(protocol.LogCmdId(cmdId))...)
					this.Conn.Write(msg1)
					this.Conn.Wait()
					this.Conn.Close()
					return
				}

				if ret != nil {
					var body []byte
					body, err = json.Marshal(ret)
					if err != nil {
						log.Error(err.Error())
						return
					}
					// var out bytes.Buffer
					// binary.Write(&out, binary.LittleEndian, uint16(0))
					// binary.Write(&out, binary.LittleEndian, msg.TransId)
					// binary.Write(&out, binary.LittleEndian, msg.CmdId)
					// binary.Write(&out, binary.LittleEndian, body)

					// data = out.Bytes()
					// binary.LittleEndian.PutUint16(data[0:2], uint16(out.Len()))

					hs := &protocol.HandShake{
						Data: body,
					}
					data, _ = protocol.MarshalMessage(msg.TransId, hs, this.Protocol)

				}
				_, err = this.Conn.Write(data)
				if err != nil {
					log.Error(err.Error())
					this.Conn.Close()
					return
				}
				this.Verified = true
				continue
			}
			if cmdId == protocol.TAG_Ping {
				//ping:=msg.Body.(*Protocol.Ping)
				//log.Info("receive ping",zap.Int64("client_session_id",this.GetId().Int64()))
				pong := &protocol.Pong{Time: time.Now()}
				data, err = protocol.MarshalMessage(msg.TransId, pong, this.Protocol)
				if err != nil {
					log.Error(err.Error())
					this.Conn.Wait()
					this.Conn.Close()
					return
				}
				_, err = this.Conn.Write(data)
				//log.Info("send ping",zap.Int64("client_session_id",this.GetId().Int64()))
				if err != nil {
					log.Error(err.Error())
					this.Conn.Close()
					return
				}
				continue
			}
			if this.ClientMsgHandler != nil {
				this.ClientMsgHandler(msg)
			} else {
				log.Error("no msg handler found")
			}
		case <-this.doneClient:
			this.closeSession()
			return
		}
	}
}

func (this *PassiveSession) inLoop() {
	for {
		select {
		case rMsg := <-this.MsgChan:
			this.OnMessage(rMsg)
		case <-this.doneServer:
			return
		}
	}
}

func (this *PassiveSession) OnMessage(msg *protocol.RouteMessage) {
	msg = this.HookReceive(msg)
	if msg != nil {
		err := this.HandleMsg(msg.FromActor, msg.TransId, msg.Body)
		if err != nil {
			log.Error("Actor:OnMessage:Error",
				lokas.LogActorReceiveMsgInfo(this, msg.Body, msg.TransId, msg.FromActor).
					Append(flog.Error(err))...,
			)
		}
	}
}

func (this *PassiveSession) closeSession() {
	if this.Manager != nil {
		this.Manager.RemoveSession(this.GetId())
	}
}

func (this *PassiveSession) stop() {
	defer func() {
		if r := recover(); r != nil {
			util.Recover(r, false)
		}
	}()
	if this.doneClient != nil {
		this.doneClient <- struct{}{}
		this.doneServer <- struct{}{}
	}
}

func (this *PassiveSession) OnOpen(conn lokas.IConn) {
	this.StartMessagePump()
	log.Info("PassiveSession:OnOpen", lokas.LogActorInfo(this)...)
	if this.Manager != nil {
		this.Manager.AddSession(this.GetId(), this)
	}
	if this.OnOpenFunc != nil {
		this.OnOpenFunc(conn)
	}
}

func (this *PassiveSession) OnClose(conn lokas.IConn) {
	if this.Manager != nil {
		this.Manager.RemoveSession(this.GetId())
	}
	log.Info("PassiveSession:OnClose", lokas.LogActorInfo(this)...)
	if this.OnCloseFunc != nil {
		this.OnCloseFunc(conn)
	}
	this.GetProcess().RemoveActor(this)
	this.stop()
}

func (this *PassiveSession) OnRecv(conn lokas.IConn, data []byte) {
	d := make([]byte, len(data), len(data))
	copy(d, data)
	this.Messages <- d
}
