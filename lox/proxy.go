package lox

import (
	"encoding/json"
	"errors"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/lox/flog"
	"github.com/nomos/go-lokas/network/tcp"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"github.com/nomos/go-lokas/util/promise"
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
		dialerCloseChans: map[util.ProcessId]chan struct{}{},
		Sessions:         NewProxySessionManager(true),
	}
	return ret
}

var _ lokas.IModule = (*Proxy)(nil)
var _ lokas.IProxy = (*Proxy)(nil)

type Proxy struct {
	Host               string
	Port               string
	server             lokas.Server
	started            bool
	mu                 sync.Mutex

	dialerCloseChans map[util.ProcessId]chan struct{}
	process         lokas.IProcess
	Sessions        *ProxySessionManager
}

func (this *Proxy) GetProcess() lokas.IProcess {
	return this.process
}

func (this *Proxy) SetProcess(process lokas.IProcess) {
	this.process = process
}

func NewProxy(process lokas.IProcess)*Proxy{
	ret := &Proxy{
		dialerCloseChans: map[util.ProcessId]chan struct{}{},
		Sessions:         NewProxySessionManager(true),
	}
	ret.process = process
	return ret
}

func (this *Proxy) OnStart() error {
	return nil
}

func (this *Proxy) OnStop() error {
	return nil
}

func (this *Proxy) Type() string {
	return ProxyCtor.Type()
}

func (this *Proxy) OnCreate() error {
	panic("implement me")
}

func (this *Proxy) OnDestroy() error {
	panic("implement me")
}


func (this *Proxy) Load(conf lokas.IConfig) error {

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

type processHandShake struct {
	Id util.ID
}

func activeSessionCreator(id util.ID,p *Proxy) func(conn lokas.IConn) lokas.ISession {
	return func(conn lokas.IConn) lokas.ISession {
		sess := NewProxySession(conn, id, p.Sessions,false)
		sess.AuthFunc = func(data []byte) error {
			sess.Verified = true
			p.Sessions.AddSession(sess.GetId(), sess)
			sess.OnVerified(true)
			return nil
		}
		sess.Protocol = protocol.BINARY
		sess.Conn = conn
		return sess
	}
}

func passiveSessionCreator(p *Proxy) func(conn lokas.IConn) lokas.ISession {
	return func(conn lokas.IConn) lokas.ISession {
		sess := NewProxySession(conn, p.GetProcess().GenId(), p.Sessions,true)
		sess.AuthFunc = func(data []byte) error {
			var hs processHandShake
			err:=json.Unmarshal(data,hs)
			if err != nil {
				log.Error(err.Error())
				return err
			}
			p.Sessions.AddSession(hs.Id, sess)
			data,_=protocol.MarshalMessage(0,hs,protocol.BINARY)
			_,err=sess.Conn.Write(data)
			if err != nil {
				log.Error(err.Error())
				sess.Conn.Close()
				sess.closeSession()
				sess.OnVerified(false)
				return err
			}
			sess.OnVerified(true)
			return nil
		}
		sess.Protocol = protocol.BINARY
		sess.Conn = conn
		return sess
	}
}

func getIdMutexKey(a,b util.ProcessId)string{
	ret:= "proxy/"
	if a>b {
		ret+=a.String()
		ret+="_"
		ret+=b.String()
	} else {
		ret+=b.String()
		ret+="_"
		ret+=a.String()
	}
	return ret
}

func (this *Proxy) checkIsConnected(id util.ProcessId)bool{
	return this.Sessions.GetSession(id.Snowflake())!=nil||this.Sessions.GetSession(id.Snowflake())!=nil
}

func (this *Proxy) connect(id util.ProcessId,addr string) (*ProxySession,error) {
	selfId:=this.GetProcess().PId()
	mu,err:=this.GetProcess().GlobalMutex(getIdMutexKey(selfId,id),15)
	if err != nil {
		log.Error(err.Error())
		return nil,err
	}
	mu.Lock()
	defer mu.Unlock()
	if this.checkIsConnected(id) {
		//如果连上
		log.Warnf("服务器已经连接",selfId.String(),id.String())
		return this.getProxySession(id),nil
	}
	context := &lokas.Context{
		SessionCreator:    activeSessionCreator(id.Snowflake(), this),
		Splitter:          protocol.Split,
		ReadBufferSize:    1024 * 1024 * 4,
		ChanSize:          10000,
		LongPacketPicker:  protocol.PickLongPacket(protocol.BINARY),
		LongPacketCreator: protocol.CreateLongPacket(protocol.BINARY),
		MaxPacketWriteLen: protocol.DEFAULT_PACKET_LEN,
	}
	//如果没有连接,尝试连接
	conn,err := tcp.Dial(addr, context)
	if err != nil {
		log.Error(err.Error())
		return nil,err
	}
	if conn==nil {
		log.Error("create session failed")
		return nil,errors.New("create session failed")
	}
	//握手协议
	activeSession:=conn.Session.(*ProxySession)
	_,err=promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		timeout:=promise.SetTimeout(time.Second*14, func(timeout *promise.Timeout) {
			reject("connect to server timeout:"+id.String())
			activeSession.Conn.Close()
			activeSession.closeSession()
			return
		})
		activeSession.OnVerified = func(success bool) {
			timeout.Close()
			if success {
				resolve(nil)
			} else {
				reject("connect to server failed:"+id.String())
			}
		}
	}).Await()
	if err != nil {
		log.Error(err.Error())
		return nil,err
	}
	return activeSession,nil
}

func (this *Proxy) getProxySession(id util.ProcessId)*ProxySession {
	sess:=this.Sessions.GetSession(id.Snowflake())
	if sess!=nil {
		return sess.(*ProxySession)
	}
	return nil
}

func (this *Proxy) Send(id util.ProcessId,msg *protocol.RouteMessage)error{
	sess:=this.getProxySession(id)
	if sess == nil {
		//info:=this.GetProcess().GetProcessInfo
		//sess,err:=this.connect(id,addr)
		//if err != nil {
		//	log.Error(err.Error())ss
		//	return err
		//}
		//TODO
		return nil
	}
	data,_:=msg.Marshal()
	_,err := sess.Conn.Write(data)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Proxy) SetPort(p string){
	this.Port = p
}

func (this *Proxy) Start() error {
	log.Info("start",flog.FuncInfo(this,"Start")...)
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

func (this *Proxy) Stop() error {
	this.mu.Lock()
	defer this.mu.Unlock()
	log.Warn("stop",flog.FuncInfo(this,"Stop")...)
	this.Sessions.Clear()
	this.started = false
	this.server.Stop()
	return nil
}
