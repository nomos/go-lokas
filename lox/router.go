package lox

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/lox/flog"
	"github.com/nomos/go-lokas/protocol"
)

var _ lokas.IModule = (*Router)(nil)
var _ lokas.IRouter = (*Router)(nil)

func NewRouter(process lokas.IProcess) *Router {
	ret := &Router{
		process:        process,
	}
	return ret
}
//router读取配置表,同步服务器注册信息
//路由信息到本机或调用Proxy建立连接
type Router struct {
	process lokas.IProcess
}

func (this *Router) GetProcess() lokas.IProcess {
	return this.process
}

func (this *Router) SetProcess(process lokas.IProcess) {
	this.process = process
}

func (this *Router) RouteMsg(msg *protocol.RouteMessage) {
	if msg.ToActor.IsValidProcessId() {
		log.Info("Router:isProcessId",flog.ActorRouterMsgInfo(msg.Body,msg.TransId,msg.FromActor,msg.ToActor)...)
	} else if msg.ToActor!=0 {
		a:=this.process.GetActor(msg.ToActor)
		log.Info("Router:RouteMsg",flog.ActorRouterMsgInfo(msg.Body,msg.TransId,msg.FromActor,msg.ToActor)...)
		if a!=nil {
			a.ReceiveMessage(msg)
			return
		}
		log.Infof("proxy msg",msg.FromActor,msg.ToActor)
		//TODO:proxy msg
		pid,err:=this.process.GetProcessIdByActor(msg.ToActor)
		log.Infof("processId",pid)

		if err != nil {
			log.Warnf("actor not found",msg.ToActor)
			return
		}
		err=this.process.Get("Proxy").(lokas.IProxy).Send(pid,msg)
		if err != nil {
			log.Error(err.Error())
			if !msg.Req {
				origin:=this.GetProcess().GetActor(msg.FromActor)
				origin.ReceiveMessage(protocol.NewRouteMessage(msg.ToActor,msg.FromActor,msg.TransId,protocol.ERR_RPC_FAILED.NewErrMsg(),true))
			}
			return
		}
	} else if msg.ToActor==0 {

	} else {
		panic("wrong id format"+msg.ToActor.String())
	}
}

func (this *Router) Type() string {
	return "Router"
}

func (this *Router) Load(conf lokas.IConfig) error {
	return nil
}

func (this *Router) Unload() error {
	return nil
}

func (this *Router) Start() error {
	return nil
}

func (this *Router) Stop() error {
	return nil
}

func (this *Router) OnStart() error {
	return nil
}

func (this *Router) OnStop() error {
	return nil
}