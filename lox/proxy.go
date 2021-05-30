package lox

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/network"
	"github.com/nomos/go-lokas/network/tcp"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"github.com/nomos/promise"
	"time"
)

var ProxyCtor = proxyCtor{}

type proxyCtor struct{}

func (this proxyCtor) Type() string {
	return "Proxy"
}

func (this proxyCtor) Create() lokas.IModule {
	ret := &Proxy{
		Actor:NewActor(),
		dialerCloseChans: map[util.ProcessId]chan struct{}{},
		ISessionManager: network.NewDefaultSessionManager(true),
	}
	return ret
}

var _ lokas.IModule = (*Proxy)(nil)
var _ lokas.IProxy = (*Proxy)(nil)


type Proxy struct {
	*Actor
	dialerCloseChans map[util.ProcessId]chan struct{}
	lokas.ISessionManager
}

func (this *Proxy) OnStart() error{
	panic("implement me")
}

func (this *Proxy) OnStop() error{
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

func (this *Proxy) Load(conf lokas.IConfig) error {
	return nil
}

func (this *Proxy) Unload() error {
	return nil
}

func sessionCreator(id util.ProcessId, p *Proxy) func(conn lokas.IConn) lokas.ISession {
	return func(conn lokas.IConn) lokas.ISession {
		sess := p.GetSession(id.Snowflake()).(*network.DefaultSession)
		if sess == nil {
			sess = network.NewDefaultSession(conn, id.Snowflake(), p)
		}
		sess.Conn = conn
		return sess
	}
}

func (this *Proxy) Connect(id util.ProcessId, addr string) error {
	context := &lokas.Context{
		SessionCreator: sessionCreator(id, this),
		Splitter:       protocol.Split,
		ReadBufferSize: 1024 * 1024 * 4,
		ChanSize:       10000,
		LongPacketPicker:  protocol.PickLongPacket(protocol.BINARY),
		LongPacketCreator: protocol.CreateLongPacket(protocol.BINARY),
		MaxPacketWriteLen: protocol.DEFAULT_PACKET_LEN,
	}
	dialerCloseChans := tcp.DialEx(addr, context, time.Second*3)
	this.dialerCloseChans[id] = dialerCloseChans
	return nil
}

func (this *Proxy) Start() *promise.Promise {
	return promise.Resolve(nil)
}

func (this *Proxy) Stop() *promise.Promise {
	this.Range(func(id util.ID, session lokas.ISession) bool {
		conn := session.GetConn()
		if conn != nil {
			conn.Close()
			conn.Wait()
		}
		return false
	})
	for _,v:=range this.dialerCloseChans {
		close(v)
	}
	return promise.Resolve(nil)
}