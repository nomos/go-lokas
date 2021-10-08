package lox

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/lox/flog"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"go.uber.org/zap"
	"time"
)

type PassiveSessionOption func(*PassiveSession)

var _ lokas.ISession = &PassiveSession{}
var _ lokas.IActor = &PassiveSession{}

func NewPassiveSession(conn lokas.IConn, id util.ID, manager lokas.ISessionManager, opts ...PassiveSessionOption) *PassiveSession {
	s := &PassiveSession{
		Actor:    NewActor(),
		Messages: make(chan []byte, 100),
		Conn:     conn,
		manager:  manager,
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
	manager          lokas.ISessionManager
	done             chan struct{}
	OnCloseFunc      func(conn lokas.IConn)
	OnOpenFunc       func(conn lokas.IConn)
	ClientMsgHandler func(msg *protocol.BinaryMessage)
	AuthFunc         func(data []byte) error
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
	log.Info("PassiveSession:StartMessagePump", flog.ActorInfo(this)...)

	this.msgChan = make(chan *protocol.RouteMessage, 100)
	this.done = make(chan struct{})
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
		//ClientSideLoop
	CLIENT_LOOP:
		for {
			select {
			case <-this.ticker.C:
				if this.OnUpdateFunc != nil && this.Verified {
					this.OnUpdateFunc()
				}
				//ClientSide MessageLoop
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
					log.Errorf("Auth Failed", cmdId)
					this.Conn.Write(msg)
					this.Conn.Wait()
					this.Conn.Close()
					return
				}
				msg, err := protocol.UnmarshalMessage(data, this.Protocol)
				if err != nil {
					log.Error("unmarshal client message error",
						zap.Any("cmdId", cmdId),
					)
					msg, _ := protocol.NewError(protocol.ERR_MSG_FORMAT).Marshal()
					_, err := this.Conn.Write(msg)
					if err != nil {
						log.Error(err.Error())
					}
					this.Conn.Close()
					break CLIENT_LOOP
				}
				if cmdId == protocol.TAG_HandShake {
					var err error
					if this.AuthFunc != nil {
						err = this.AuthFunc(msg.Body.(*protocol.HandShake).Data)
					}
					if err != nil {
						log.Error(err.Error())
						msg, err := protocol.MarshalMessage(msg.TransId, protocol.NewError(protocol.ERR_AUTH_FAILED), this.Protocol)
						if err != nil {
							log.Error(err.Error())
							this.Conn.Wait()
							this.Conn.Close()
							return
						}
						log.Errorf("Auth Failed", cmdId)
						this.Conn.Write(msg)
						this.Conn.Wait()
						this.Conn.Close()
						break CLIENT_LOOP
					}
					_, err = this.Conn.Write(data)
					if err != nil {
						log.Error(err.Error())
						this.Conn.Close()
						break CLIENT_LOOP
					}
					this.Verified = true
					continue
				}
				if cmdId == protocol.TAG_Ping {
					//ping:=msg.Body.(*Protocol.Ping)
					pong := &protocol.Pong{Time: time.Now()}
					data, err := protocol.MarshalMessage(msg.TransId, pong, this.Protocol)
					if err != nil {
						log.Error(err.Error())
						this.Conn.Wait()
						this.Conn.Close()
						break CLIENT_LOOP
					}
					_, err = this.Conn.Write(data)
					if err != nil {
						log.Error(err.Error())
						this.Conn.Close()
						break CLIENT_LOOP
					}
					continue
				}
				if this.ClientMsgHandler != nil {
					this.ClientMsgHandler(msg)
				} else {
					log.Error("no msg handler found")
				}
			case <-this.done:
				this.closeSession()
				//log.Warnf("done")
				break CLIENT_LOOP
			}
		}
		close(this.done)
		close(this.Messages)
	}()

	go func() {
		defer func() {
			r := recover()
			if r != nil {
				if util.Recover(r,true)!=nil {
					log.Error("服务端协议出错")
					this.Conn.Close()
				}
			}
		}()
		SERVER_LOOP:
		for {
			select {
			//ServerSideMsgLoop
			case rMsg := <-this.msgChan:
				this.OnMessage(rMsg)
			case <-this.done:
				break SERVER_LOOP
			}
		}
		close(this.msgChan)
	}()
}

func (this *PassiveSession) closeSession() {
	if this.manager != nil {
		this.manager.RemoveSession(this.GetId())
	}
}

func (this *PassiveSession) stop() {
	this.done<- struct{}{}
}

func (this *PassiveSession) OnOpen(conn lokas.IConn) {
	this.StartMessagePump()
	log.Info("PassiveSession:OnOpen", flog.ActorInfo(this)...)
	if this.manager != nil {
		this.manager.AddSession(this.GetId(), this)
	}
	if this.OnOpenFunc != nil {
		this.OnOpenFunc(conn)
	}
}

func (this *PassiveSession) OnClose(conn lokas.IConn) {
	if this.manager != nil {
		this.manager.RemoveSession(this.GetId())
	}
	log.Info("PassiveSession:OnClose", flog.ActorInfo(this)...)
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
