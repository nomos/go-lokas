package lox

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/nomos/go-lokas/log/flog"
	"regexp"
	"strconv"
	"time"

	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

var _ lokas.IModule = &Registry{}
var _ lokas.IRegistry = &Registry{}

type Registry struct {
	process               lokas.IProcess
	LocalRegistry         *CommonRegistry //local actor&service registry
	GlobalRegistry        *CommonRegistry //local actor&service registry
	actorWatchCloseChan   chan struct{}
	processWatchCloseChan chan struct{}
	serviceWatchCloseChan chan struct{}

	serviceRegisterMgr *ServiceRegisterMgr
	serviceDiscoverMgr *ServiceDiscoverMgr

	timer   *time.Ticker
	done    chan struct{}
	leaseId clientv3.LeaseID
}

func NewRegistry(process lokas.IProcess) *Registry {
	ret := &Registry{
		process:            process,
		LocalRegistry:      NewCommonRegistry(),
		GlobalRegistry:     NewCommonRegistry(),
		serviceRegisterMgr: NewServiceRegisterMgr(process),
		serviceDiscoverMgr: NewServiceDiscoverMgr(process),
	}
	return ret
}

func (reg *Registry) GetServiceRegisterMgr() lokas.IServiceRegisterMgr {
	return reg.serviceRegisterMgr
}

func (reg *Registry) GetServiceDiscoverMgr() lokas.IServiceDiscoverMgr {
	return reg.serviceDiscoverMgr
}

func (this *Registry) GetActorIdsByTypeAndServerId(serverId int32, typ string) []util.ID {
	if serverId == this.GetProcess().ServerId() {
		log.Warn("GetLocalServer", flog.ServerId(serverId), zap.Int32("self_server_id", this.GetProcess().ServerId()))
		return this.LocalRegistry.GetActorIdsByTypeAndServerId(serverId, typ)
	}
	return this.GlobalRegistry.GetActorIdsByTypeAndServerId(serverId, typ)
}

func (this *Registry) GetProcessInfo() string {
	return ""
}

func (this *Registry) GetProcessIdByActor(actorId util.ID) (util.ProcessId, error) {
	regi := this.GlobalRegistry.GetActorRegistry(actorId)
	if regi == nil {
		return 0, protocol.ERR_ACTOR_NOT_FOUND
	}
	return regi.ProcessId, nil
}

func (this *Registry) OnCreate() error {
	panic("implement me")
}

func (this *Registry) OnDestroy() error {
	panic("implement me")
}

//return leaseId,(bool)is registered,error
func (this *Registry) GetLeaseId() (clientv3.LeaseID, bool, error) {
	c := this.process.GetEtcd()
	if this.leaseId != 0 {
		resToLive, err := c.Lease.TimeToLive(context.Background(), this.leaseId)
		if err != nil {
			log.Error(err.Error())
			return 0, false, err
		}
		//if lease id is expired,create a new lease id
		if resToLive.TTL <= 0 {
			res, err := c.Lease.Grant(context.Background(), LeaseDuration)
			if err != nil {
				log.Error(err.Error())
				return 0, false, err
			}
			this.leaseId = res.ID
			return this.leaseId, false, nil
		}
		if resToLive.TTL < LeaseRenewDuration {
			_, err := c.Lease.KeepAliveOnce(context.Background(), this.leaseId)
			if err != nil {
				log.Error(err.Error())
				return 0, false, err
			}
		}
		return this.leaseId, true, nil
	}

	res, err := c.Lease.Grant(context.Background(), LeaseDuration)
	if err != nil {
		log.Error(err.Error())
		return 0, false, err
	}
	this.leaseId = res.ID
	return this.leaseId, false, nil
}

func (this *Registry) update() {
	this.updateProcessInfo()
}

func (this *Registry) start() {
	this.registerProcessInfo()
	this.timer = time.NewTicker(time.Second * 5)
	this.done = make(chan struct{})
	go func() {
	LOOP:
		for {
			select {
			case <-this.timer.C:
				this.update()
			case <-this.done:
				break LOOP
			}
		}
		close(this.done)
		this.unregisterProcessInfo()
	}()
}

func (this *Registry) Start() error {
	this.start()
	return nil
}

func (this *Registry) Stop() error {
	if this.done != nil {
		this.done <- struct{}{}
	}
	this.OnStop()
	return nil
}

func (this *Registry) OnStart() error {
	log.Info("Registry:OnStart")
	return nil
}

func (this *Registry) OnStop() error {
	log.Info("Registry:OnStop")

	this.serviceRegisterMgr.Stop()
	this.serviceDiscoverMgr.Stop()

	return nil
}

func (this *Registry) GetProcess() lokas.IProcess {
	return this.process
}

func (this *Registry) SetProcess(process lokas.IProcess) {
	this.process = process
}

func (this *Registry) Type() string {
	return "Registry"
}

func (this *Registry) Load(conf lokas.IConfig) error {
	this.startUpdateRemoteActorInfo()
	this.startUpdateRemoteProcessInfo()
	// err := this.startUpdateRemoteService()
	err := this.serviceDiscoverMgr.StartDiscover()
	if err != nil {
		return err
	}

	return nil
}

func (this *Registry) Unload() error {
	this.actorWatchCloseChan <- struct{}{}
	this.processWatchCloseChan <- struct{}{}
	return nil
}

//check if actor key exist,otherwise add it
func (this *Registry) checkOrCreateActorRegistry(kv *mvccpb.KeyValue) {
	id, _ := strconv.Atoi(regexp.MustCompile(`[/]actor[/]([0-9]+)`).ReplaceAllString(string(kv.Key), "$1"))
	actorId := util.ID(id)
	actorReg := NewActorRegistry(actorId)
	json.Unmarshal(kv.Value, actorReg)
	// log.Warn("add actor registry success", flog.FuncInfo(this, "checkOrCreateActorRegistry").Append(flog.Result(log.PrettyStruct(actorReg)))...)

	log.Debug("add actor registry success", zap.Uint64("actor", uint64(actorId)), zap.Any("reg", actorReg))
	this.GlobalRegistry.AddActor(actorReg)
}

func (this *Registry) deleteActorRegistry(kv *mvccpb.KeyValue) {
	id, _ := strconv.Atoi(regexp.MustCompile(`[/]actor[/]([0-9]+)`).ReplaceAllString(string(kv.Key), "$1"))
	actorId := util.ID(id)
	this.GlobalRegistry.RemoveActor(actorId)
}

//check if process key exist,otherwise add it
func (this *Registry) checkOrCreateProcessRegistry(kv *mvccpb.KeyValue) {
	id, _ := strconv.Atoi(regexp.MustCompile(`[/]processids[/]([0-9]+)`).ReplaceAllString(string(kv.Key), "$1"))
	// log.Debug("checkOrCreateProcessRegistry", zap.Int("id", id))

	pid := util.ProcessId(id)
	processReg := NewProcessRegistry(pid)
	this.GlobalRegistry.AddProcess(processReg)

	log.Debug("add process registry success", zap.Uint16("pid", uint16(pid)), zap.Any("reg", processReg))
}

func (this *Registry) deleteProcessRegistry(kv *mvccpb.KeyValue) {
	id, _ := strconv.Atoi(regexp.MustCompile(`[/]processids[/]([0-9]+)`).ReplaceAllString(string(kv.Key), "$1"))
	pid := util.ProcessId(id)
	this.GlobalRegistry.RemoveProcess(pid)
}

//update actor registries via etcd
func (this *Registry) startUpdateRemoteActorInfo() error {
	log.Info("start", flog.FuncInfo(this, "startUpdateRemoteActorInfo")...)
	client := this.GetProcess().GetEtcd()
	res, err := client.Get(context.TODO(), "/actor/", clientv3.WithPrefix())

	if err != nil {
		log.Error(err.Error())
		return err
	}
	for _, v := range res.Kvs {
		this.checkOrCreateActorRegistry(v)
	}
	watcher := client.Watch(context.TODO(), "/actor/", clientv3.WithPrefix())
	this.actorWatchCloseChan = make(chan struct{})
	go func() {
	LOOP:
		for {
			select {
			case resp := <-watcher:
				for _, e := range resp.Events {
					if e.Type == mvccpb.PUT {
						//log.Warn("PUT actor",
						//	flog.FuncInfo(this, "startUpdateRemoteActorInfo").
						//		Concat(flog.KeyValue(string(e.Kv.Key), string(e.Kv.Value)))...,
						//)
						this.checkOrCreateActorRegistry(e.Kv)
					} else if e.Type == mvccpb.DELETE {
						//log.Warn("DELETE actor",
						//	flog.FuncInfo(this, "startUpdateRemoteActorInfo").
						//		Concat(flog.KeyValue(string(e.Kv.Key), string(e.Kv.Value)))...
						//)
						this.deleteActorRegistry(e.Kv)
					}
				}
			case <-this.actorWatchCloseChan:
				break LOOP
			}
		}
		close(this.actorWatchCloseChan)
	}()
	return nil
}

//update process registries information via etcd
func (this *Registry) startUpdateRemoteProcessInfo() error {
	log.Info("start", flog.FuncInfo(this, "startUpdateRemoteProcessInfo")...)
	client := this.GetProcess().GetEtcd()
	res, err := client.Get(context.TODO(), "/processids/", clientv3.WithPrefix())
	if err != nil {
		log.Error(err.Error())
		return err
	}
	for _, v := range res.Kvs {
		this.checkOrCreateProcessRegistry(v)
	}
	watchChan := client.Watch(context.TODO(), "/processids/", clientv3.WithPrefix(), clientv3.WithRev(res.Header.Revision))
	this.processWatchCloseChan = make(chan struct{})
	go func() {
	LOOP:
		for {
			select {
			case resp := <-watchChan:
				for _, e := range resp.Events {
					if e.Type == mvccpb.PUT {
						//log.Warn("PUT Process Registry",
						//	flog.FuncInfo(this, "startUpdateRemoteProcessInfo").
						//		Concat(flog.KeyValue(string(e.Kv.Key), string(e.Kv.Value)))...
						//)
						this.checkOrCreateProcessRegistry(e.Kv)
					} else if e.Type == mvccpb.DELETE {
						//log.Warn("DELETE Process Registry",
						//	flog.FuncInfo(this, "startUpdateRemoteProcessInfo").
						//		Concat(flog.KeyValue(string(e.Kv.Key), string(e.Kv.Value)))...
						//)
						this.deleteProcessRegistry(e.Kv)
					}
				}
			case <-this.processWatchCloseChan:
				break LOOP
			}
		}
		close(this.processWatchCloseChan)
	}()
	return nil
}

func (this *Registry) addServiceFromEtcd(kv *mvccpb.KeyValue) error {

	serviceInfo := &ServiceRegistry{}
	err := json.Unmarshal(kv.Value, serviceInfo)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	this.GlobalRegistry.AddService(serviceInfo)
	log.Infof("update service", log.PrettyStruct(serviceInfo))
	return nil
}

func (this *Registry) delServiceFromEtcd(kv *mvccpb.KeyValue) error {
	// serviceInfo := &ServiceRegistry{}
	// err := json.Unmarshal(kv.Value, serviceInfo)
	// if err != nil {
	// 	log.Error(err.Error())
	// 	return err
	// }

	reg := regexp.MustCompile("/service/(?P<type>[a-zA-Z]+)/(?P<id>[0-9]+)")
	matchs := reg.FindStringSubmatch(string(kv.Key))
	typIdx := reg.SubexpIndex("type")
	idIdx := reg.SubexpIndex("id")

	if len(matchs) < idIdx+1 || len(matchs) < typIdx+1 || typIdx < 0 || idIdx < 0 {
		err := errors.New("etc data invalid")
		log.Error(err.Error())
		return err
	}
	serviceType := matchs[typIdx]
	serviceId, err := strconv.ParseUint(matchs[idIdx], 10, 16)
	if err != nil {
		log.Error(err.Error(), zap.String("etcd://service/kv.Key", string(kv.Key)))
		return err
	}

	this.GlobalRegistry.RemoveService(serviceType, uint16(serviceId))
	log.Info("del service", flog.ServiceInfo(serviceType, uint16(serviceId), 0)...)
	return nil
}

func (this *Registry) startUpdateRemoteService() error {
	log.Info("start", flog.FuncInfo(this, "startUpdateRemoteService")...)
	etcdClient := this.GetProcess().GetEtcd()
	resp, err := etcdClient.Get(context.TODO(), "/service/", clientv3.WithPrefix())
	if err != nil {
		log.Error(err.Error())
		return err
	}

	watchChan := etcdClient.Watch(context.TODO(), "/service/", clientv3.WithPrefix(), clientv3.WithRev(resp.Header.Revision))

	for _, v := range resp.Kvs {
		this.addServiceFromEtcd(v)
	}

	this.serviceWatchCloseChan = make(chan struct{})
	go func() {
	LOOP:
		for {
			select {
			case resp := <-watchChan:
				for _, v := range resp.Events {
					switch v.Type {
					case mvccpb.PUT:
						this.addServiceFromEtcd(v.Kv)
					case mvccpb.DELETE:
						this.delServiceFromEtcd(v.Kv)
					}
				}
			case <-this.serviceWatchCloseChan:
				break LOOP
			}
		}
		close(this.serviceWatchCloseChan)
	}()

	return nil
}

func (this *Registry) updateProcessInfo() error {
	client := this.GetProcess().GetEtcd()
	leaseId, isReg, err := this.GetLeaseId()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if !isReg {
		res, err := client.Put(context.TODO(), "/processids/"+this.process.PId().ToString()+"", time.Now().String(), clientv3.WithLease(leaseId))
		if err != nil {
			log.Error(err.Error())
			return err
		}
		log.Warnf("res", res)
	}
	return nil
}

func (this *Registry) unregisterProcessInfo() error {
	client := this.GetProcess().GetEtcd()
	leaseId, _, err := this.GetLeaseId()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	_, err = client.Lease.Revoke(context.TODO(), leaseId)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Registry) registerProcessInfo() error {
	// log.Info("registerProcessInfo")
	client := this.GetProcess().GetEtcd()
	s, err := json.Marshal(CreateProcessRegistryInfo(this.GetProcess()))
	if err != nil {
		log.Error(err.Error())
		return err
	}

	_, err = client.Put(context.TODO(), "/process/"+this.process.PId().ToString()+"/info", string(s))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Info("register process info complete", zap.Uint16("pid", uint16(this.process.GetId())))

	return nil
}

func (this *Registry) RegisterActors() error {
	client := this.GetProcess().GetEtcd()
	s, err := json.Marshal(CreateProcessActorsInfo(this.GetProcess()))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	_, err = client.Put(context.TODO(), "/process/"+this.process.PId().ToString()+"/actors", string(s))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	// log.Debug("res", flog.FuncInfo(this, "RegisterActors").Append(flog.Result(res.Header.String()))...)
	return nil
}

func (this *Registry) RegisterActorRemote(actor lokas.IActor) error {
	client := this.GetProcess().GetEtcd()
	s, err := json.Marshal(CreateActorRegistryInfo(actor))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	prefix := "/actor/"
	leaseId, isReg, err := actor.GetLeaseId()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if !isReg {
		res, err := client.Put(context.TODO(), prefix+actor.GetId().String(), string(s), clientv3.WithLease(leaseId))
		if err != nil {
			log.Error(err.Error())
			return err
		}

		// arr := flog.FuncInfo(this, "RegisterActorRemote").Append(flog.Result(res.Header.String()))
		log.Debug("register actor remote", lokas.LogActorInfo(actor).Append(flog.Result(res.Header.String()))...)
	}
	return nil
}

func (this *Registry) UnregisterActorRemote(actor lokas.IActor) error {
	client := this.GetProcess().GetEtcd()
	leaseId, _, err := actor.GetLeaseId()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	_, err = client.Lease.Revoke(context.TODO(), leaseId)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Registry) QueryRemoteActorsByType(typ string) []*Actor {
	ret := []*Actor{}
	return ret
}

func (this *Registry) QueryRemoteActorsByServer(typ string, ServerId int32) []*Actor {
	ret := []*Actor{}
	return ret
}

func (this *Registry) RegisterActorLocal(actor lokas.IActor) error {
	// log.Debug("register",
	// 	flog.FuncInfo(this, "RegisterActorLocal").
	// 		Concat(flog.ActorInfo(actor))...,
	// )
	re := &ActorRegistry{
		Id:        actor.GetId(),
		Type:      actor.Type(),
		ProcessId: this.process.PId(),
		GameId:    this.process.GameId(),
		Version:   this.process.Version(),
		ServerId:  this.process.ServerId(),
		Ts:        time.Now(),
	}
	this.LocalRegistry.AddActor(re)
	return nil
}

func (this *Registry) UnregisterActorLocal(actor lokas.IActor) error {
	this.LocalRegistry.RemoveActor(actor.GetId())
	return nil
}
