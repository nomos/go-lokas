package lox

import (
	"context"
	"sync"

	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/lox/flog"
	"github.com/nomos/go-lokas/network"
	"github.com/nomos/go-lokas/network/tcp"
	"github.com/nomos/go-lokas/network/ws"
	"github.com/nomos/go-lokas/protocol"
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

func (this gateCtor) Create() lokas.IModule {

	ctx, cancel := context.WithCancel(context.Background())
	ret := &Gate{
		Actor:           NewActor(),
		ISessionManager: network.NewDefaultSessionManager(true),
		Ctx:             ctx,
		Cancel:          cancel,
	}
	ret.SetType("Gate")
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
	started            bool
	mu                 sync.Mutex
	Ctx                context.Context
	Cancel             context.CancelFunc
}

func (this *Gate) OnStart() error {
	return nil
}

func (this *Gate) OnStop() error {
	return nil
}

func (this *Gate) LoadCustom(host, port string, protocolType protocol.TYPE, connType ConnType) error {
	this.Host = host
	this.Port = port
	this.Protocol = protocolType
	this.connType = connType
	sessionFunc := this.SessionCreator
	if this.SessionCreatorFunc != nil {
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

	return this.LoadCustom(conf.GetString("host"), conf.GetString("port"), protocol.String2Type(conf.GetString("protocol")), String2ConnType(conf.GetString("conn")))
}

func (this *Gate) SessionCreator(conn lokas.IConn) lokas.ISession {
	sess := NewPassiveSession(conn, this.GetProcess().GenId(), this)
	sess.AuthFunc = this.AuthFunc
	sess.Protocol = this.Protocol
	this.ISessionManager.AddSession(sess.GetId(), sess)
	this.GetProcess().AddActor(sess)
	this.GetProcess().StartActor(sess)
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

func (this *Gate) Start() error {
	log.Info(this.Type() + " Start")
	this.mu.Lock()
	defer this.mu.Unlock()
	if this.started {
		return nil
	}
	err := this.server.Start("0.0.0.0" + ":" + this.Port)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	this.started = true
	return nil
}
func (this *Gate) Stop() error {
	this.mu.Lock()
	defer this.mu.Unlock()

	this.started = false
	this.Cancel()
	this.ISessionManager.Clear()
	log.Warn("stop", flog.FuncInfo(this, "Stop")...)

	this.server.Stop()
	return nil
}

func (this *Gate) GetServer() lokas.Server {
	return this.server
}
