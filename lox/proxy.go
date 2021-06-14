package lox

import (
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/network"
	"github.com/nomos/go-lokas/network/tcp"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"github.com/nomos/promise"
	"sync"
	"time"
)

var ProxyCtor = proxyCtor{}

type proxyCtor struct{}

func (this proxyCtor) Type() string {
	return "Proxy"
}

func (this proxyCtor) Create() lokas.IModule {
	ret := &Proxy{
		Actor:            NewActor(),
		dialerCloseChans: map[util.ProcessId]chan struct{}{},
		ActiveSessions:   network.NewDefaultSessionManager(true),
		PassiveSessions:   network.NewDefaultSessionManager(true),
	}
	return ret
}

var _ lokas.IModule = (*Proxy)(nil)
var _ lokas.IProxy = (*Proxy)(nil)

type Proxy struct {
	*Actor
	Host               string
	Port               string
	server             lokas.Server
	startPending       *promise.Promise
	started            bool
	mu                 sync.Mutex

	dialerCloseChans map[util.ProcessId]chan struct{}
	ActiveSessions   lokas.ISessionManager
	PassiveSessions  lokas.ISessionManager
}

func (this *Proxy) OnStart() error {
	panic("implement me")
}

func (this *Proxy) OnStop() error {
	panic("implement me")
}

func (this *Proxy) Type() string {
	return "Proxy"
}

func (this *Proxy) OnCreate() error {
	panic("implement me")
}

func (this *Proxy) OnDestroy() error {
	panic("implement me")
}

func passiveSessionCreator(p *Proxy) func(conn lokas.IConn) lokas.ISession {
	return func(conn lokas.IConn) lokas.ISession {
		sess := NewPassiveSession(conn, p.GetProcess().GenId(), p.PassiveSessions)
		//sess.AuthFunc = this.AuthFunc
		sess.Protocol = protocol.BINARY
		p.PassiveSessions.AddSession(sess.GetId(), sess)
		sess.Conn = conn
		return sess
	}
}

func (this *Proxy) Load(conf lokas.IConfig) error {
	log.WithFields(log.Fields{
		"host":     conf.Get("host"),
		"port":     conf.Get("port"),
		"protocol": conf.Get("protocol"),
		"conn":     conf.Get("conn"),
	}).Info("Gate:LoadConfig")
	this.Host = conf.Get("host").(string)
	this.Port = conf.Get("port").(string)
	context := &lokas.Context{
		SessionCreator:    passiveSessionCreator(this),
		Splitter:          protocol.Split,
		ReadBufferSize:    1024 * 1024,
		ChanSize:          200,
		LongPacketPicker:  protocol.PickLongPacket(protocol.BINARY),
		LongPacketCreator: protocol.CreateLongPacket(protocol.BINARY),
		MaxPacketWriteLen: protocol.DEFAULT_PACKET_LEN,
	}
	this.server = tcp.NewServer(context)
	return nil
}

func (this *Proxy) Unload() error {
	return nil
}

func activeSessionCreator(id util.ProcessId, p *Proxy) func(conn lokas.IConn) lokas.ISession {
	return func(conn lokas.IConn) lokas.ISession {
		sess := p.ActiveSessions.GetSession(id.Snowflake()).(*network.DefaultSession)
		if sess == nil {
			sess = network.NewDefaultSession(conn, id.Snowflake(), p.ActiveSessions)
		}
		sess.Conn = conn
		return sess
	}
}

func (this *Proxy) Connect(id util.ProcessId, addr string) error {
	context := &lokas.Context{
		SessionCreator:    activeSessionCreator(id, this),
		Splitter:          protocol.Split,
		ReadBufferSize:    1024 * 1024 * 4,
		ChanSize:          10000,
		LongPacketPicker:  protocol.PickLongPacket(protocol.BINARY),
		LongPacketCreator: protocol.CreateLongPacket(protocol.BINARY),
		MaxPacketWriteLen: protocol.DEFAULT_PACKET_LEN,
	}
	dialerCloseChans := tcp.DialEx(addr, context, time.Second*3)
	this.dialerCloseChans[id] = dialerCloseChans
	return nil
}

func (this *Proxy) Start() *promise.Promise {
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

func (this *Proxy) Stop() *promise.Promise {
	this.mu.Lock()
	defer this.mu.Unlock()
	log.Warnf("Proxy:Stop")
	this.ActiveSessions.Clear()
	for _, v := range this.dialerCloseChans {
		close(v)
	}
	this.PassiveSessions.Clear()
	this.started = false
	this.server.Stop()
	return promise.Resolve(nil)
}
