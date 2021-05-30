package lox

import (
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/network"
	"github.com/nomos/go-lokas/network/tcp"
	"github.com/nomos/go-lokas/network/ws"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/promise"
	"sync"
)

type ConnType int

const (
	TCP       ConnType = 0
	Websocket ConnType = 1
)

func String2ConnType(s string) ConnType {
	switch s {
	case "tcp":
		return TCP
	case "ws":
		return Websocket
	default:
		panic("not a valid conn type")
	}
}

func (this ConnType) String() string {
	switch this {
	case TCP:
		return "tcp"
	case Websocket:
		return "ws"
	default:
		panic("not a valid conn type")
	}
}

var GateCtor = gateCtor{}

type gateCtor struct{}

func (this gateCtor) Type() string {
	return "Gate"
}

func (this gateCtor) Create() lokas.IModule {
	ret := &Gate{
		Actor:NewActor(),
		ISessionManager: network.NewDefaultSessionManager(true),
	}
	return ret
}

var _ lokas.IActor = (*Gate)(nil)

type Gate struct {
	*Actor
	lokas.ISessionManager
	Host               string
	Port               string
	AuthFunc           func(data []byte) error
	SessionCreatorFunc func(conn lokas.IConn) lokas.ISession
	Protocol           protocol.TYPE
	connType           ConnType
	server             lokas.Server
	startPending       *promise.Promise
	started            bool
	mu                 sync.Mutex
}

func (this *Gate) OnStart() error{
	return nil
}

func (this *Gate) OnStop() error{
	return nil
}

func (this *Gate) Type() string {
	return GateCtor.Type()
}

func (this *Gate) Load(conf lokas.IConfig) error {
	log.WithFields(log.Fields{
		"host":     conf.Get("host"),
		"port":     conf.Get("port"),
		"protocol": conf.Get("protocol"),
		"conn":     conf.Get("conn"),
	}).Info("Gate:LoadConfig")
	this.Host = conf.Get("host").(string)
	this.Port = conf.Get("port").(string)
	this.Protocol = protocol.String2Type(conf.Get("protocol").(string))
	this.connType = String2ConnType(conf.Get("conn").(string))
	sessionFunc := this.SessionCreator
	if this.SessionCreatorFunc!= nil {
		sessionFunc = this.SessionCreatorFunc
	}
	if this.connType == Websocket {
		log.Info("creating ws gate on " + this.Protocol.String() + " Protocol")
		context := &lokas.Context{
			SessionCreator:    sessionFunc,
			Splitter:          protocol.Split,
			ChanSize:          200,
			LongPacketPicker:  protocol.PickLongPacket(this.Protocol),
			LongPacketCreator: protocol.CreateLongPacket(this.Protocol),
			MaxPacketWriteLen: protocol.DEFAULT_PACKET_LEN,
		}
		this.server = ws.NewWsServer(context)
	}
	if this.connType == TCP {
		log.Info("creating tcp gate on " + this.Protocol.String() + " Protocol")
		context := &lokas.Context{
			SessionCreator:    sessionFunc,
			Splitter:          protocol.Split,
			ReadBufferSize:    1024 * 1024,
			ChanSize:          200,
			LongPacketPicker:  protocol.PickLongPacket(this.Protocol),
			LongPacketCreator: protocol.CreateLongPacket(this.Protocol),
			MaxPacketWriteLen: protocol.DEFAULT_PACKET_LEN,
		}
		this.server = tcp.NewServer(context)
	}
	return nil
}

func (this *Gate) SessionCreator(conn lokas.IConn) lokas.ISession {
	sess := NewClientSession(conn, this.GetProcess().GenId(), this,
		WithAuthFunc(this.AuthFunc),
		WithProtocol(this.Protocol),
	)
	this.ISessionManager.AddSession(sess.GetId(), sess)
	this.GetProcess().AddActor(sess)
	return sess
}

func (this *Gate) Unload() error {
	return nil
}

func (this *Gate) OnCreate() error {
	log.Info(this.Type() + " OnCreate")
	return nil
}

func (this *Gate) OnDestroy() error {
	log.Info(this.Type() + " OnDestroy")
	return nil
}

func (this *Gate) Start() *promise.Promise {
	log.Info(this.Type() + " Start")
	this.mu.Lock()
	defer this.mu.Unlock()
	if this.startPending == nil && !this.started {
		this.startPending = promise.Async(func(resolve func(interface{}), reject func(interface{})) {
			err := this.server.Start(this.Host + ":" + this.Port)
			this.mu.Lock()
			defer this.mu.Unlock()
			if err != nil {
				this.startPending = nil
				reject(err)
				return
			}
			this.started = true
			this.startPending = nil
			resolve(nil)
		})
	} else if this.started {
		return promise.Resolve(nil)
	}
	return this.startPending
}
func (this *Gate) Stop() *promise.Promise {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.ISessionManager.Clear()
	log.Warnf("Gate:Stop")
	this.started = false
	this.server.Stop()
	return promise.Resolve(nil)
}
