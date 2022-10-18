package lox

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/lox/flog"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"go.uber.org/zap"
)

var _ lokas.IModule = (*Router)(nil)
var _ lokas.IRouter = (*Router)(nil)

func NewRouter(process lokas.IProcess) *Router {
	ret := &Router{
		process: process,
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
		log.Info("Router:isProcessId", flog.ActorRouterMsgInfo(msg.Body, msg.TransId, msg.FromActor, msg.ToActor, msg.Req)...)
	} else if msg.ToActor != 0 {
		a := this.process.GetActor(msg.ToActor)
		log.Info("Router:RouteMsg", flog.ActorRouterMsgInfo(msg.Body, msg.TransId, msg.FromActor, msg.ToActor, msg.Req)...)
		if a != nil {
			a.ReceiveMessage(msg)
			return
		}

		pid, err := this.process.GetProcessIdByActor(msg.ToActor)
		// log.Info("processId", flog.ProcessId(pid.Snowflake()))
		if err != nil {
			log.Error("actor not found", zap.Uint64("toActorId", uint64(msg.ToActor)), zap.Uint16("cmd", uint16(msg.InnerId)))
			return
		}

		err = this.process.Send(pid, msg)
		if err != nil {
			log.Error(err.Error())
			if !msg.Req {
				// TODO RPC MSG
				origin := this.GetProcess().GetActor(msg.FromActor)
				origin.ReceiveMessage(protocol.NewRouteMessage(msg.ToActor, msg.FromActor, msg.TransId, protocol.ERR_RPC_FAILED.NewErrMsg(), true))
			}
			return
		}
	} else if msg.ToActor == 0 {

	} else {
		panic("wrong id format" + msg.ToActor.String())
	}
}

func (router *Router) RouteMsgLocal(msg *protocol.RouteMessage) error {
	a := router.process.GetActor(msg.ToActor)
	log.Info("Router:RouteMsgLocal", flog.ActorRouterMsgInfo(msg.Body, msg.TransId, msg.FromActor, msg.ToActor, msg.Req)...)
	if a == nil {
		log.Error("route local, not find actor", zap.Any("routeMsg", msg))
		return protocol.ERR_MSG_ROUTE_NOT_FOUND
	}

	a.ReceiveMessage(msg)
	return nil
}

func (router *Router) RouteMsgByAvatar(fromActorId util.ID, toActorId util.ID, transId uint32, reqType uint8, msg protocol.ISerializable) error {
	cmd, _ := msg.GetId()
	routeMsg := protocol.RouteMessage{
		TransId:   transId,
		ReqType:   reqType,
		CmdId:     cmd,
		InnerId:   cmd,
		FromActor: fromActorId,
		ToActor:   toActorId,
		Body:      msg,
	}

	if routeMsg.ReqType == protocol.REQ_TYPE_REPLAY {
		routeMsg.Req = false
	} else {
		routeMsg.Req = true
	}

	a := router.process.GetActor(routeMsg.ToActor)
	if a != nil {
		a.ReceiveMessage(&routeMsg)
		return nil
	}

	// TODO send by proxy

	return nil
}

func (router *Router) RouteMsgByService(fromActorId util.ID, serviceType string, serviceId uint16, lineId uint16, transId uint32, reqType uint8, msg protocol.ISerializable, protocolType protocol.TYPE) error {

	serviceInfo, ok := router.GetProcess().GetServiceDiscoverMgr().FindServiceInfo(serviceType, serviceId, lineId)
	if !ok {
		log.Error("route msg err, not find service", zap.String("serviceType", serviceType), zap.Uint16("serviceId", serviceId), zap.Uint16("lineId", lineId))
		return protocol.ERR_INTERNAL_SERVER
	}

	cmd, _ := msg.GetId()
	routeMsg := protocol.RouteMessage{
		TransId:   transId,
		ReqType:   reqType,
		CmdId:     cmd,
		InnerId:   cmd,
		FromActor: fromActorId,
		ToActor:   serviceInfo.ActorId,
		Body:      msg,
	}

	if router.GetProcess().PId() == serviceInfo.ProcessId {

		a := router.process.GetActor(routeMsg.ToActor)
		if a == nil {
			log.Warn("router local process, not find actor", zap.Uint16("pid", uint16(serviceInfo.ProcessId)), zap.Uint64("toActor", uint64(routeMsg.ToActor)))
			return protocol.ERR_ACTOR_NOT_FOUND
		}
		log.Info("router local actor", zap.Uint16("toPid", uint16(serviceInfo.ProcessId)), zap.Uint64("fromActorId", uint64(fromActorId)), zap.Uint64("toActorId", uint64(routeMsg.ToActor)), zap.Uint16("cmd", uint16(cmd)))

		if routeMsg.ReqType == protocol.REQ_TYPE_REPLAY {
			routeMsg.Req = false
		} else {
			routeMsg.Req = true
		}

		a.ReceiveMessage(&routeMsg)
		return nil
	} else {
		// remote
		outData, err := protocol.MarshalRouteMsg(&routeMsg, protocolType)
		if err != nil {
			log.Error("marsh route msg err", zap.Any("routeMsg", routeMsg))
		}

		router.GetProcess().SendData(serviceInfo.ProcessId, outData)
		log.Info("router send routeMsg", zap.String("serviceType", serviceType), zap.Uint16("serviceId", serviceId), zap.Uint16("lineId", lineId), zap.Uint16("toPid", uint16(serviceInfo.ActorId.ProcessId())), zap.Uint64("fromActor", uint64(fromActorId)), zap.Uint64("toActor", uint64(routeMsg.ToActor)), zap.Uint16("cmd", uint16(cmd)))
	}

	return nil
}

func (router *Router) RouteDataByService(routeDataMsg *protocol.RouteDataMsg, serviceType string, serviceId uint16, lineId uint16) error {

	serviceInfo, ok := router.GetProcess().GetServiceDiscoverMgr().FindServiceInfo(serviceType, serviceId, lineId)
	if !ok {
		log.Error("route data msg err, not find service", zap.String("serviceType", serviceType), zap.Uint16("serviceId", serviceId), zap.Uint16("lineId", lineId))
		return protocol.ERR_INTERNAL_SERVER
	}

	outData, err := routeDataMsg.MarshalData()
	if err != nil {
		return err
	}

	if router.GetProcess().PId() == serviceInfo.ProcessId {
		a := router.process.GetActor(routeDataMsg.ToActor)
		if a == nil {
			log.Error("router local process, not find actor", zap.Uint16("pid", uint16(serviceInfo.ProcessId)), zap.Uint64("toActor", uint64(routeDataMsg.ToActor)))
			return protocol.ERR_ACTOR_NOT_FOUND
		}
		// log.Info("router local actor", zap.Uint16("toPid", uint16(serviceInfo.ProcessId)), zap.Uint64("fromActorId", uint64(fromActorId)), zap.Uint64("toActorId", uint64(routeMsg.ToActor)), zap.Uint16("cmd", uint16(cmd)))
		// a.ReceiveMessage(&routeDataMsg)
		recv := &protocol.RouteRecv{
			Protocol: routeDataMsg.Protocol,
			Data:     outData,
		}
		a.ReceiveData(recv)
		return nil
	} else {
		// remote
		// router.GetProcess().Send(serviceInfo.ProcessId, &routeMsg)
		// log.Info("router send routeMsg", zap.Uint16("toPid", uint16(serviceInfo.ActorId.ProcessId())), zap.Uint64("fromActorId", uint64(fromActorId)), zap.Uint64("toActorId", uint64(routeMsg.ToActor)), zap.Uint16("cmd", uint16(cmd)))

		router.GetProcess().SendData(serviceInfo.ProcessId, outData)
	}

	return nil
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
