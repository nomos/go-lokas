package lox

import (
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"github.com/nomos/promise"
	"go.uber.org/zap"
	"time"
)

var _ lokas.ISession = &ClientSession{}
var _ lokas.IActor = &ClientSession{}

func NewClientSession(conn lokas.IConn, id util.ID, manager lokas.ISessionManager, opts ...SessionOption) *ClientSession {
	s := &ClientSession{
		Actor:NewActor(),
		Messages: make(chan []byte, 100),
		Conn:     conn,
		manager:  manager,
		timeout: TimeOut,
		ticker: time.NewTicker(UpdateTime),
	}
	s.SetId(id)
	for _, o := range opts {
		o(s)
	}
	return s
}

type ClientSession struct {
	*Actor
	Verified    bool
	Messages    chan []byte
	Conn        lokas.IConn
	Protocol protocol.TYPE
	manager     lokas.ISessionManager
	done        chan struct{}
	OnCloseFunc func(conn lokas.IConn)
	OnOpenFunc  func(conn lokas.IConn)
	ClientMsgHandler func(msg *protocol.BinaryMessage)
	AuthFunc    func(data []byte) error
	timeout     time.Duration
	ticker *time.Ticker
}

func (this *ClientSession) Load(conf lokas.IConfig) error {
	panic("implement me")
}

func (this *ClientSession) Unload() error {
	panic("implement me")
}

func (this *ClientSession) OnStart() error {
	panic("implement me")
}

func (this *ClientSession) OnStop() error {
	panic("implement me")
}

type SessionOption func(*ClientSession)

func WithCloseFunc(closeFunc func(conn lokas.IConn)) SessionOption {
	return func(session *ClientSession) {
		session.OnCloseFunc = closeFunc
	}
}

func WithTimeout(timeout time.Duration) SessionOption {
	return func(session *ClientSession) {
		session.timeout = timeout
	}
}

func WithOpenFunc(openFunc func(conn lokas.IConn)) SessionOption {
	return func(session *ClientSession) {
		session.OnCloseFunc = openFunc
	}
}

func WithAuthFunc(authFunc func(data []byte) error) SessionOption {
	return func(session *ClientSession) {
		session.AuthFunc = authFunc
	}
}

func WithProtocol(protocol protocol.TYPE) SessionOption {
	return func(session *ClientSession) {
		session.Protocol = protocol
	}
}

func (this *ClientSession) Type() string {
	return "ClientSession"
}

func (this *ClientSession) OnCreate() error {
	return nil
}

func (this *ClientSession) Start() *promise.Promise {
	return promise.Resolve(nil)
}

func (this *ClientSession) Stop() *promise.Promise {
	return promise.Resolve(nil)
}

func (this *ClientSession) OnDestroy() error {
	return nil
}

func (this *ClientSession) GetConn() lokas.IConn {
	return this.Conn
}

func (this *ClientSession) StartMessagePump() {
	log.Infof("StartMessagePump message pump",this.genId())
	this.msgChan = make(chan *protocol.RouteMessage,100)
	this.done = make(chan struct{})
	go func() {
		for {
			select {
			case <-this.ticker.C:
				if this.OnUpdateFunc!=nil&&this.Verified {
					this.OnUpdateFunc()
				}
			case data := <-this.Messages:
				cmdId := protocol.GetCmdId16(data)
				if !this.Verified && cmdId != protocol.TAG_HandShake {
					var msg []byte
					msg,err:=protocol.MarshalMessage(0,protocol.NewError(protocol.ErrAuthFailed),this.Protocol)
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
				msg, err := protocol.UnmarshalMessage(data,this.Protocol)
				if err != nil {
					log.Error("unmarshal client message error",
						zap.Any("cmdId", cmdId),
					)
					msg, _ := protocol.NewError(protocol.ErrMsgFormat).Marshal()
					_,err:=this.Conn.Write(msg)
					if err != nil {
						log.Error(err.Error())
					}
					this.Conn.Close()
					return
				}
				if cmdId == protocol.TAG_HandShake {
					var err error
					if this.AuthFunc!= nil {
						err = this.AuthFunc(msg.Body.(*protocol.HandShake).Data)
					}
					if err != nil {
						log.Error(err.Error())
						msg,err:=protocol.MarshalMessage(msg.TransId,protocol.NewError(protocol.ErrAuthFailed),this.Protocol)
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
					pong:=&protocol.Pong{Time: time.Now()}
					data,err:=protocol.MarshalMessage(msg.TransId,pong,this.Protocol)
					if err != nil {
						log.Error(err.Error())
						this.Conn.Wait()
						this.Conn.Close()
						return
					}
					_,err =this.Conn.Write(data)
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
			case <-this.done:
				this.closeSession()
				//log.Warnf("done")
				return
			}
		}
	}()
	go func() {
		for {
			select {
			case rMsg := <-this.msgChan:
				this.OnMessage(rMsg)
			case <-this.done:
				return
			}
		}
	}()
}

func (this *ClientSession) closeSession() {
	if this.manager != nil {
		this.manager.RemoveSession(this.GetId())
	}
}

func (this *ClientSession) stop() {
	close(this.done)
}

func (this *ClientSession) OnOpen(conn lokas.IConn) {
	this.StartMessagePump()
	if this.OnOpenFunc != nil {
		this.OnOpenFunc(conn)
	}
	log.Infof("ClientSession:OnOpen")
	if this.manager != nil {
		this.manager.AddSession(this.GetId(), this)
	}
}

func (this *ClientSession) OnClose(conn lokas.IConn) {
	if this.manager != nil {
		this.manager.RemoveSession(this.GetId())
	}
	log.Infof("ClientSession:OnClose")
	if this.OnCloseFunc != nil {
		this.OnCloseFunc(conn)
	}
	this.GetProcess().RemoveActor(this)
	this.stop()
}

func (this *ClientSession) OnRecv(conn lokas.IConn, data []byte) {
	d := make([]byte, len(data), len(data))
	copy(d, data)
	this.Messages <- d
}
