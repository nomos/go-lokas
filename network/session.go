package network

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/ecs"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"go.uber.org/zap"
	"time"
)

var _ lokas.ISession = &DefaultSession{}

func NewDefaultSession(conn lokas.IConn, id util.ID, manager lokas.ISessionManager,opts ...SessionOption) *DefaultSession {
	s := &DefaultSession{
		ID:id,
		Messages: make(chan []byte, 100),
		Conn:     conn,
		manager:  manager,
		done:     make(chan struct{}),
	}
	for _,o:=range opts {
		o(s)
	}
	return s
}


type SessionOption func(*DefaultSession)

func WithCloseFunc(closeFunc func(conn lokas.IConn)) SessionOption {
	return func(session *DefaultSession) {
		session.OnCloseFunc = closeFunc
	}
}

func WithOpenFunc(closeFunc func(conn lokas.IConn)) SessionOption {
	return func(session *DefaultSession) {
		session.OnCloseFunc = closeFunc
	}
}

func WithMsgHandler(msgHandler func(msg *protocol.BinaryMessage)) SessionOption {
	return func(session *DefaultSession) {
		session.MsgHandler = msgHandler
	}
}

type DefaultSession struct {
	util.ID
	process     lokas.IProcess
	Messages    chan []byte
	Conn        lokas.IConn
	manager     lokas.ISessionManager
	done        chan struct{}
	OnCloseFunc func(conn lokas.IConn)
	OnOpenFunc  func(conn lokas.IConn)
	MsgHandler  func(msg *protocol.BinaryMessage)
}

func (this *DefaultSession) OnMessage(actorId util.ID, transId uint32, msg protocol.ISerializable) error {
	panic("implement me")
}

func (this *DefaultSession) SendMessage(actorId util.ID, transId uint32, msg protocol.ISerializable) error {
	panic("implement me")
}

func (this *DefaultSession) Call(actorId util.ID, transId uint32, req protocol.ISerializable, resp protocol.ISerializable) error {
	panic("implement me")
}

func (this *DefaultSession) AsyncCall(actorId util.ID, transId uint32, req protocol.ISerializable, resp protocol.ISerializable) error {
	panic("implement me")
}

func (this *DefaultSession) Type() string {
	panic("implement me")
}

func (this *DefaultSession) SetId(id util.ID) {
	this.ID = id
}

func (this *DefaultSession) CloneEntity() *ecs.Entity {
	panic("implement me")
}

func (this *DefaultSession) Update(dt time.Duration, now time.Time) {
	panic("implement me")
}

func (this *DefaultSession) GetProcess() lokas.IProcess {
	panic("implement me")
}

func (this *DefaultSession) SetProcess(process lokas.IProcess) {
	panic("implement me")
}

func (this *DefaultSession) OnCreate() error {
	panic("implement me")
}

func (this *DefaultSession) Start() error {
	this.start()
	return nil
}

func (this *DefaultSession) Stop() error {
	panic("implement me")
}

func (this *DefaultSession) OnDestroy() error {
	panic("implement me")
}

func (this *DefaultSession) GetId()util.ID {
	return this.ID
}

func (this *DefaultSession) GetConn() lokas.IConn {
	return this.Conn
}

func (this *DefaultSession) OnOpen(conn lokas.IConn) {
	this.start()
	if this.OnOpenFunc !=nil {
		this.OnOpenFunc(conn)
	}
	log.Warnf("OnOpen")
	if this.manager!=nil {
		this.manager.AddSession(this.ID, this)
	}
}

func (this *DefaultSession) OnClose(conn lokas.IConn) {
	if this.manager!=nil {
		this.manager.RemoveSession(this.ID)
	}
	log.Warnf("OnClose")
	if this.OnOpenFunc != nil {
		this.OnCloseFunc(conn)
	}
	this.stop()
}

func (this *DefaultSession) closeSession(){
	if this.manager!=nil {
		this.manager.RemoveSession(this.ID)
	}
}

func (this *DefaultSession) Write(data []byte)error{
	_, err := this.Conn.Write(data)
	return err
}

func (this *DefaultSession) OnRecv(conn lokas.IConn, data []byte) {
	// 注意: 此处data直接引用的网络缓冲区的slice，如果把data发送给其他goroutine处理，需要注意缓冲区覆盖问题
	data1 := make([]byte, len(data), len(data))
	copy(data1, data)
	this.Messages <- data1
}

func (this *DefaultSession) handleMsg(msg *protocol.BinaryMessage) {
	if this.MsgHandler !=nil {
		this.MsgHandler(msg)
	}
}

func (this *DefaultSession) start() {
	go func() {
		for {
			select {
			case data := <-this.Messages:
				cmdId := protocol.GetCmdId16(data)
				msg, err := protocol.UnmarshalBinaryMessage(data)
				if err != nil {
					log.Error("unmarshal client message error",
						zap.Any("cmdId", cmdId),
					)
					this.Conn.Close()
					return
				}
				//if err != nil {
				//	log.Error("route client message to server error, CmdId: %d, error: %s",
				//		zap.Any("cmdId", cmdId),
				//		zap.Any("err", err),
				//	)
				//	this.IConn.Close()
				//}
				this.handleMsg(msg)
			case <-this.done:
				this.closeSession()
			}
		}
	}()
}

func (this *DefaultSession) stop(){
	this.done<- struct{}{}
}

func (this *DefaultSession) HandleMessage(f func(msg *protocol.BinaryMessage)) {
	this.MsgHandler = f
}
