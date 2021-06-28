package lox

import (
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/ecs"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"go.uber.org/zap"
	"time"
)

type SessionOption func(*ActiveSession)

var _ lokas.ISession = &ActiveSession{}

func NewActiveSession(conn lokas.IConn, id util.ID, manager lokas.ISessionManager, opts ...SessionOption) *ActiveSession {
	s := &ActiveSession{
		ID:       id,
		Messages: make(chan []byte, 100),
		Conn:     conn,
		manager:  manager,
		done:     make(chan struct{}),
		timeout:  time.Second * 15,
		Protocol: protocol.BINARY,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

type ActiveSession struct {
	util.ID
	process     lokas.IProcess
	Messages    chan []byte
	Conn        lokas.IConn
	Protocol 	protocol.TYPE
	manager     lokas.ISessionManager
	done        chan struct{}
	OnCloseFunc func(conn lokas.IConn)
	OnOpenFunc  func(conn lokas.IConn)
	MsgHandler  func(msg *protocol.BinaryMessage)
	timeout     time.Duration
	pingIndex   uint32
}

func (this *ActiveSession) OnMessage(actorId util.ID, transId uint32, msg protocol.ISerializable) error {
	panic("implement me")
}

func (this *ActiveSession) SendMessage(actorId util.ID, transId uint32, msg protocol.ISerializable) error {
	panic("implement me")
}

func (this *ActiveSession) Call(actorId util.ID, transId uint32, req protocol.ISerializable, resp protocol.ISerializable) error {
	panic("implement me")
}

func (this *ActiveSession) AsyncCall(actorId util.ID, transId uint32, req protocol.ISerializable, resp protocol.ISerializable) error {
	panic("implement me")
}

func (this *ActiveSession) Type() string {
	panic("implement me")
}

func (this *ActiveSession) SetId(id util.ID) {
	this.ID = id
}

func (this *ActiveSession) CloneEntity() *ecs.Entity {
	panic("implement me")
}

func (this *ActiveSession) Update(dt time.Duration, now time.Time) {
	panic("implement me")
}

func (this *ActiveSession) GetProcess() lokas.IProcess {
	panic("implement me")
}

func (this *ActiveSession) SetProcess(process lokas.IProcess) {
	panic("implement me")
}

func (this *ActiveSession) OnCreate() error {
	panic("implement me")
}

func (this *ActiveSession) Start() error {
	this.start()
	return nil
}

func (this *ActiveSession) Stop() error {
	panic("implement me")
}

func (this *ActiveSession) OnDestroy() error {
	panic("implement me")
}

func (this *ActiveSession) GetId() util.ID {
	return this.ID
}

func (this *ActiveSession) GetConn() lokas.IConn {
	return this.Conn
}

func (this *ActiveSession) OnOpen(conn lokas.IConn) {
	this.start()
	if this.OnOpenFunc != nil {
		this.OnOpenFunc(conn)
	}
	log.Warnf("OnOpen")
	if this.manager != nil {
		this.manager.AddSession(this.ID, this)
	}
}

func (this *ActiveSession) OnClose(conn lokas.IConn) {
	if this.manager != nil {
		this.manager.RemoveSession(this.ID)
	}
	log.Warnf("OnClose")
	if this.OnOpenFunc != nil {
		this.OnCloseFunc(conn)
	}
	this.stop()
}

func (this *ActiveSession) closeSession() {
	if this.manager != nil {
		this.manager.RemoveSession(this.ID)
	}
}

func (this *ActiveSession) Write(data []byte) error {
	_, err := this.Conn.Write(data)
	return err
}

func (this *ActiveSession) OnRecv(conn lokas.IConn, data []byte) {
	// 注意: 此处data直接引用的网络缓冲区的slice，如果把data发送给其他goroutine处理，需要注意缓冲区覆盖问题
	data1 := make([]byte, len(data), len(data))
	copy(data1, data)
	this.Messages <- data1
}

func (this *ActiveSession) handleMsg(msg *protocol.BinaryMessage) {
	if this.MsgHandler != nil {
		this.MsgHandler(msg)
	}
}

func (this *ActiveSession) PongHandler(pong *protocol.Pong) {
	this.Conn.SetReadDeadline(time.Now().Add(this.timeout))
}

func (this *ActiveSession) start() {
	go func() {
		ticker := time.NewTicker(this.timeout / 5)
		defer func() {
			ticker.Stop()
			this.Conn.Close()
		}()
		this.Conn.SetReadDeadline(time.Now().Add(this.timeout))
	Loop:
		for {
			select {
			case <-ticker.C:
				ping := &protocol.Ping{Time: time.Now()}
				this.pingIndex++
				data, _ := protocol.MarshalMessage(this.pingIndex, ping,this.Protocol)
				_, err := this.Conn.Write(data)
				if err != nil {
					log.Error(err.Error())
					return
				}
			case data := <-this.Messages:
				cmdId := protocol.GetCmdId16(data)
				msg, err := protocol.UnmarshalMessage(data,this.Protocol)
				if err != nil {
					log.Error("unmarshal client message error",
						zap.Any("cmdId", cmdId),
					)
					return
				}
				if cmdId == protocol.TAG_Pong {
					this.PongHandler(msg.Body.(*protocol.Pong))
					continue
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
				log.Warnf("closing")
				this.closeSession()
				break Loop
			}
		}
	}()
}

func (this *ActiveSession) stop() {
	this.done <- struct{}{}
}

func (this *ActiveSession) HandleMessage(f func(msg *protocol.BinaryMessage)) {
	this.MsgHandler = f
}
