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

var _ lokas.ISession = (*ProxySession)(nil)
var _ lokas.IActor = (*PassiveSession)(nil)

type SessionOption func(*ProxySession)

func WithCloseFunc(closeFunc func(conn lokas.IConn)) SessionOption {
	return func(session *ProxySession) {
		session.OnCloseFunc = closeFunc
	}
}

func WithOpenFunc(closeFunc func(conn lokas.IConn)) SessionOption {
	return func(session *ProxySession) {
		session.OnCloseFunc = closeFunc
	}
}

func WithMsgHandler(msgHandler func(msg *protocol.BinaryMessage)) SessionOption {
	return func(session *ProxySession) {
		session.MsgHandler = msgHandler
	}
}

func NewProxySession(conn lokas.IConn, id util.ID, manager *ProxySessionManager,passive bool, opts ...SessionOption) *ProxySession {
	s := &ProxySession{
		Actor:NewActor(),
		Messages: make(chan []byte, 100),
		Conn:     conn,
		passive: passive,
		manager:  manager,
		timeout: TimeOut,
		ticker: time.NewTicker(UpdateTime),
	}
	for _, o := range opts {
		o(s)
	}
	s.SetType("ProxySession")
	s.SetId(id)
	return s
}

type ProxySession struct {
	*Actor
	Verified    bool
	Messages    chan []byte
	Conn        lokas.IConn
	Protocol protocol.TYPE
	passive     bool				//是否为被动连接
	manager     *ProxySessionManager
	done        chan struct{}
	OnCloseFunc func(conn lokas.IConn)
	OnOpenFunc  func(conn lokas.IConn)
	OnVerified 	func(success bool)
	MsgHandler  func(msg *protocol.BinaryMessage)
	AuthFunc    func(data []byte) error
	timeout     time.Duration
	pingIndex   uint32
	ticker *time.Ticker
}

func (this *ProxySession) SendMessage(actorId util.ID, transId uint32, msg protocol.ISerializable) error {
	panic("implement me")
}

func (this *ProxySession) Type() string {
	panic("implement me")
}

func (this *ProxySession) Update(dt time.Duration, now time.Time) {
	panic("implement me")
}

func (this *ProxySession) GetProcess() lokas.IProcess {
	panic("implement me")
}

func (this *ProxySession) SetProcess(process lokas.IProcess) {
	panic("implement me")
}

func (this *ProxySession) OnCreate() error {
	panic("implement me")
}

func (this *ProxySession) Start() error {
	this.start()
	return nil
}

func (this *ProxySession) Stop() error {
	panic("implement me")
}

func (this *ProxySession) OnDestroy() error {
	panic("implement me")
}

func (this *ProxySession) GetConn() lokas.IConn {
	return this.Conn
}

func (this *ProxySession) OnOpen(conn lokas.IConn) {
	this.start()
	if this.OnOpenFunc != nil {
		this.OnOpenFunc(conn)
	}
	log.Warn("OnOpen")
	if this.manager != nil {
		this.manager.AddSession(this.GetId(), this)
	}
}

func (this *ProxySession) OnClose(conn lokas.IConn) {
	if this.manager != nil {
		this.manager.RemoveSession(this.GetId())
	}
	log.Warn("OnClose")
	if this.OnOpenFunc != nil {
		this.OnCloseFunc(conn)
	}
	this.stop()
}

func (this *ProxySession) closeSession() {
	if this.manager != nil {
		this.manager.RemoveSession(this.GetId())
	}
}

func (this *ProxySession) Write(data []byte) error {
	_, err := this.Conn.Write(data)
	return err
}

func (this *ProxySession) OnRecv(conn lokas.IConn, data []byte) {
	// 注意: 此处data直接引用的网络缓冲区的slice，如果把data发送给其他goroutine处理，需要注意缓冲区覆盖问题
	data1 := make([]byte, len(data), len(data))
	copy(data1, data)
	this.Messages <- data1
}

func (this *ProxySession) handleMsg(msg *protocol.BinaryMessage) {
	if this.MsgHandler != nil {
		this.MsgHandler(msg)
	}
}

func (this *ProxySession) startMessagePumpPassive(){
	this.msgChan = make(chan *protocol.RouteMessage,100)
	this.done = make(chan struct{})
	go func() {
		defer func() {
			r:=recover()
			if r!=nil {
				if e,ok:=r.(error);ok {
					log.Errorf(e.Error())
					log.Error("客户端协议出错")
					this.Conn.Close()
				}
			}
		}()
		LOOP:
			for {
				select {
				case <-this.ticker.C:
					if this.OnUpdateFunc!=nil&&this.Verified {
						this.OnUpdateFunc()
					}
				case rMsg := <-this.msgChan:
					this.OnMessage(rMsg)
				case data := <-this.Messages:
					cmdId := protocol.GetCmdId16(data)
					//第一个包必须是握手包
					if !this.Verified && cmdId != protocol.TAG_HandShake {
						msg,_:=protocol.MarshalMessage(0,protocol.NewError(protocol.ERR_AUTH_FAILED),this.Protocol)
						log.Errorf("Auth Failed", cmdId)
						_,err:=this.Conn.Write(msg)
						if err != nil {
							log.Error(err.Error())
							this.Conn.Close()
							break LOOP
						}
						this.Conn.Wait()
						this.Conn.Close()
						break LOOP
					}
					msg, err := protocol.UnmarshalMessage(data,this.Protocol)
					if err != nil {
						log.Error("unmarshal client message error",
							zap.Any("cmdId", cmdId),
						)
						msg, _ := protocol.NewError(protocol.ERR_MSG_FORMAT).Marshal()
						_,err:=this.Conn.Write(msg)
						if err != nil {
							log.Error(err.Error())
							this.Conn.Close()
							break LOOP
						}
						this.Conn.Wait()
						this.Conn.Close()
						break LOOP
					}
					if cmdId == protocol.TAG_HandShake {
						var err error
						if this.AuthFunc!= nil {
							err = this.AuthFunc(msg.Body.(*protocol.HandShake).Data)
						}
						if err != nil {
							log.Error(err.Error())
							msg,_:=protocol.MarshalMessage(msg.TransId,protocol.NewError(protocol.ERR_AUTH_FAILED),this.Protocol)
							log.Errorf("Auth Failed", cmdId)
							_,err:=this.Conn.Write(msg)
							if err != nil {
								log.Error(err.Error())
								this.Conn.Close()
								break LOOP
							}
							this.Conn.Wait()
							this.Conn.Close()
							break LOOP
						}
						_, err = this.Conn.Write(data)
						if err != nil {
							log.Error(err.Error())
							this.Conn.Close()
							break LOOP
						}
						this.Verified = true
						continue
					}
					if cmdId == protocol.TAG_Ping {
						//ping:=msg.Body.(*Protocol.Ping)
						pong:=&protocol.Pong{Time: time.Now()}
						data,err:=protocol.MarshalMessage(msg.TransId,pong,this.Protocol)
						_,err =this.Conn.Write(data)
						if err != nil {
							log.Error(err.Error())
							this.Conn.Close()
							break LOOP
						}
						continue
					}
				case <-this.done:
					this.Conn.Close()
					this.closeSession()
					break LOOP
				}
			}
	}()
}

func (this *ProxySession) startMessagePumpActive(){
	this.msgChan = make(chan *protocol.RouteMessage,100)
	this.done = make(chan struct{})
	go func() {
		defer func() {
			r:=recover()
			if r!=nil {
				if e,ok:=r.(error);ok {
					log.Errorf(e.Error())
					log.Error("客户端协议出错")
					this.Conn.Close()
				}
			}
		}()
		ticker := time.NewTicker(this.timeout / 5)
		defer func() {
			ticker.Stop()
			this.Conn.Close()
		}()
		this.Conn.SetReadDeadline(time.Now().Add(this.timeout))
	LOOP:
		for {
			select {
			case <-this.ticker.C:
				ping := &protocol.Ping{Time: time.Now()}
				this.pingIndex++
				data, _ := protocol.MarshalMessage(this.pingIndex, ping,this.Protocol)
				_, err := this.Conn.Write(data)
				if err != nil {
					log.Error(err.Error())
					return
				}
				if this.OnUpdateFunc!=nil&&this.Verified {
					this.OnUpdateFunc()
				}
			case rMsg := <-this.msgChan:
				this.OnMessage(rMsg)
			case data := <-this.Messages:
				cmdId := protocol.GetCmdId16(data)
				//第一个包必须是握手包
				if !this.Verified && cmdId != protocol.TAG_HandShake {
					msg,_:=protocol.MarshalMessage(0,protocol.NewError(protocol.ERR_AUTH_FAILED),this.Protocol)
					log.Errorf("Auth Failed", cmdId)
					_,err:=this.Conn.Write(msg)
					if err != nil {
						log.Error(err.Error())
						this.Conn.Close()
						break LOOP
					}
					this.Conn.Wait()
					this.Conn.Close()
					break LOOP
				}
				msg, err := protocol.UnmarshalMessage(data,this.Protocol)
				if err != nil {
					log.Error("unmarshal client message error",
						zap.Any("cmdId", cmdId),
					)
					msg, _ := protocol.NewError(protocol.ERR_MSG_FORMAT).Marshal()
					_,err:=this.Conn.Write(msg)
					if err != nil {
						log.Error(err.Error())
						this.Conn.Close()
						break LOOP
					}
					this.Conn.Wait()
					this.Conn.Close()
					break LOOP
				}
				if cmdId == protocol.TAG_HandShake {
					var err error
					if this.AuthFunc!= nil {
						err = this.AuthFunc(msg.Body.(*protocol.HandShake).Data)
					}
					if err != nil {
						log.Error(err.Error())
						msg,_:=protocol.MarshalMessage(msg.TransId,protocol.NewError(protocol.ERR_AUTH_FAILED),this.Protocol)
						log.Errorf("Auth Failed", cmdId)
						_,err:=this.Conn.Write(msg)
						if err != nil {
							log.Error(err.Error())
							this.Conn.Close()
							break LOOP
						}
						this.Conn.Wait()
						this.Conn.Close()
						break LOOP
					}
					_, err = this.Conn.Write(data)
					if err != nil {
						log.Error(err.Error())
						this.Conn.Close()
						break LOOP
					}
					this.Verified = true
					continue
				}
				if cmdId == protocol.TAG_Pong {
					//ping:=msg.Body.(*Protocol.Ping)
					this.Conn.SetReadDeadline(time.Now().Add(this.timeout))
					continue
				}
				this.handleMsg(msg)
			case <-this.done:
				log.Warn("closing",flog.FuncInfo(this,"start")...)
				this.Conn.Close()
				this.closeSession()
				break LOOP
			}
		}
	}()
}

func (this *ProxySession) start() {
	if this.passive {
		this.startMessagePumpPassive()
	} else {
		this.startMessagePumpActive()
	}
}

func (this *ProxySession) stop() {
	this.done <- struct{}{}
}

func (this *ProxySession) HandleMessage(f func(msg *protocol.BinaryMessage)) {
	this.MsgHandler = f
}
