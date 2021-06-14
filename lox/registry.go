package lox

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"github.com/nomos/promise"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
	"regexp"
	"strconv"
	"sync"
	"time"
)

func NewCommonRegistry() *CommonRegistry {
	ret := &CommonRegistry{
		Processes:    map[util.ProcessId]*ProcessRegistry{},
		Service:      map[protocol.BINARY_TAG]*ServiceRegistry{},
		Actors:       map[util.ID]*ActorRegistry{},
		ActorsByType: map[string][]util.ID{},
		ActorsByServer: map[int32][]util.ID{},
		Ts:           time.Time{},
	}
	return ret
}

type CommonRegistry struct {
	Processes    map[util.ProcessId]*ProcessRegistry
	Service      map[protocol.BINARY_TAG]*ServiceRegistry
	Actors       map[util.ID]*ActorRegistry
	ActorsByType map[string][]util.ID
	ActorsByServer map[int32][]util.ID
	Ts           time.Time
	mu           sync.Mutex
}

func (this *CommonRegistry)GetActorRegistry(id util.ID)*ActorRegistry{
	return this.Actors[id]
}

func (this *CommonRegistry)GetActorIdsByTypeAndServerId(serverId int32,typ string)[]util.ID{
	ret:=[]util.ID{}
	serverIds,ok:=this.ActorsByServer[serverId]
	if !ok {
		return ret
	}
	typeIds,ok:=this.ActorsByType[typ]
	if !ok {
		return ret
	}
	for _,a:=range serverIds {
		for _,b:=range typeIds {
			if a==b {
				ret = append(ret, a)
			}
		}
	}
	return ret
}

func removeIdFromArr(id util.ID,arr []util.ID)[]util.ID{
	ret:=[]util.ID{}
	for _,v:=range arr {
		if v!=id {
			ret = append(ret, id)
		}
	}
	return ret
}

func addId2ArrOnce(id util.ID,arr []util.ID)[]util.ID{
	for _,v:=range arr {
		if v==id {
			return arr
		}
	}
	arr = append(arr, id)
	return arr
}

func (this *CommonRegistry) AddActor(actor *ActorRegistry){
	this.mu.Lock()
	defer this.mu.Unlock()
	this.Actors[actor.Id] = actor
	if actorArr,ok:=this.ActorsByType[actor.Type];ok {
		actorArr = addId2ArrOnce(actor.Id,actorArr)
		this.ActorsByType[actor.Type] = actorArr
	} else {
		this.ActorsByType[actor.Type] = []util.ID{actor.Id}
	}
	if actorArr,ok:=this.ActorsByServer[actor.ServerId];ok {
		actorArr = addId2ArrOnce(actor.Id,actorArr)
		this.ActorsByServer[actor.ServerId] = actorArr
	} else {
		this.ActorsByServer[actor.ServerId] = []util.ID{actor.Id}
	}
}

func (this *CommonRegistry) RemoveActor(actorId util.ID){
	this.mu.Lock()
	defer this.mu.Unlock()
	if actor,ok:=this.Actors[actorId];ok {
		delete(this.Actors,actorId)
		if actorArr,ok:=this.ActorsByType[actor.Type];ok {
			this.ActorsByType[actor.Type] = removeIdFromArr(actor.Id,actorArr)
		}
		if actorArr,ok:=this.ActorsByServer[actor.ServerId];ok {
			this.ActorsByServer[actor.ServerId] = removeIdFromArr(actor.Id,actorArr)
		}
	}
}

func (this *CommonRegistry) AddProcess(process *ProcessRegistry){
	this.mu.Lock()
	defer this.mu.Unlock()
	this.Processes[process.Id] = process
}

func (this *CommonRegistry) RemoveProcess(id util.ProcessId){
	this.mu.Lock()
	defer this.mu.Unlock()
	delete(this.Processes,id)
}

func (this *CommonRegistry) AddService(service *ServiceRegistry){
	this.mu.Lock()
	defer this.mu.Unlock()
	this.Service[service.Id] = service
}

func (this *CommonRegistry) RemoveService(id protocol.BINARY_TAG){
	this.mu.Lock()
	defer this.mu.Unlock()
	delete(this.Service,id)
}

type ServiceRegistry struct {
	Id          protocol.BINARY_TAG
	ServiceType lokas.ServiceType
	GameId      string
	Version     string
	ServerId    uint32
	Weights     map[util.ID]int
	Ts          time.Time
}

type ActorRegistry struct {
	Id        util.ID
	ProcessId util.ProcessId
	Type      string
	GameId    string
	Version   string
	ServerId  int32
	//Health    lokas.ActorState
	Ts        time.Time
}

func NewActorRegistry(id util.ID)*ActorRegistry {
	ret:=&ActorRegistry{
		Id:        id,
		ProcessId: 0,
		Type:      "",
		GameId:    "",
		Version:   "",
		ServerId:  0,
		Ts:        time.Time{},
	}
	return ret
}

type ProcessRegistry struct {
	Id       util.ProcessId
	GameId   string
	Version  string
	ServerId int32
	Host     string
	Port     string
	Services map[protocol.BINARY_TAG]*lokas.Service
	Actors   map[util.ID]*ActorRegistry
	Ts       time.Time
}

func NewProcessRegistry(id util.ProcessId)*ProcessRegistry {
	ret:=&ProcessRegistry{
		Id:       id,
		GameId:   "",
		Version:  "",
		ServerId: 0,
		Host:     "",
		Port:     "",
		Services: nil,
		Actors:   nil,
		Ts:       time.Time{},
	}
	return ret
}

type ProcessActorsInfo struct {
	Id util.ProcessId
	Actors []util.ID
	Ts time.Time
}

func CreateProcessActorsInfo(process lokas.IProcess)*ProcessActorsInfo{
	ret:=&ProcessActorsInfo{
		Id:     process.Id(),
		Actors: process.GetActorIds(),
		Ts: time.Now(),
	}
	return ret
}

type ActorRegistryInfo struct {
	Id util.ID
	Type string
	ProcessId util.ProcessId
	GameId   string
	Version  string
	ServerId int32
	Ts       time.Time
}

func CreateActorRegistryInfo(actor lokas.IActor)*ActorRegistryInfo{
	ret:=&ActorRegistryInfo{
		Id:        actor.GetId(),
		Type:      actor.Type(),
		ProcessId: actor.GetProcess().Id(),
		GameId:    actor.GetProcess().GameId(),
		Version:   actor.GetProcess().Version(),
		ServerId:  actor.GetProcess().ServerId(),
		Ts:        time.Now(),
	}
	return ret
}

type ProcessServiceInfo struct {
	Id util.ProcessId
	Services map[protocol.BINARY_TAG]int
}

type ProcessRegistryInfo struct {
	Id util.ProcessId
	GameId   string
	Version  string
	ServerId int32
	Host     string
	Port     string
	Ts time.Time
}

func CreateProcessRegistryInfo(process lokas.IProcess)*ProcessRegistryInfo{
	ret:=&ProcessRegistryInfo{
		Id:       process.Id(),
		GameId:   process.GameId(),
		Version:  process.Version(),
		ServerId: process.ServerId(),
		Host:     process.Config().GetString("host"),
		Port:     process.Config().GetString("port"),
		Ts:       time.Now(),
	}
	return ret
}

var _ lokas.IModule = &Registry{}
var _ lokas.IRegistry = &Registry{}

func NewRegistry(process lokas.IProcess) *Registry {
	ret := &Registry{
		process:        process,
		LocalRegistry:  NewCommonRegistry(),
		GlobalRegistry: NewCommonRegistry(),
	}
	return ret
}

type Registry struct {
	process             lokas.IProcess
	LocalRegistry       *CommonRegistry //local actor&service registry
	GlobalRegistry      *CommonRegistry //local actor&service registry
	actorWatchCloseChan chan struct{}
	processWatchCloseChan chan struct{}
}

func (this *Registry) GetActorIdsByTypeAndServerId(serverId int32,typ string)[]util.ID {
	if serverId==this.GetProcess().ServerId() {
		return this.LocalRegistry.GetActorIdsByTypeAndServerId(serverId,typ)
	}
	return this.GlobalRegistry.GetActorIdsByTypeAndServerId(serverId,typ)
}

func (this *Registry) GetProcessIdByActor(actorId util.ID)(util.ProcessId,error){
	regi:=this.GlobalRegistry.GetActorRegistry(actorId)
	if regi==nil {
		return 0,protocol.ERR_ACTOR_NOT_FOUND
	}
	return regi.ProcessId,nil
}

func (this *Registry) OnCreate() error {
	panic("implement me")
}

func (this *Registry) OnDestroy() error {
	panic("implement me")
}

func (this *Registry) Start() *promise.Promise {
	return promise.Resolve(nil)
}

func (this *Registry) Stop() *promise.Promise {
	return promise.Resolve(nil)
}

func (this *Registry) OnStart() error{
	log.Infof("Registry:OnStart")
	this.RegisterProcessInfo()
	return nil
}

func (this *Registry) OnStop() error{
	log.Infof("Registry:OnStop")
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
	return nil
}

func (this *Registry) Unload() error {
	this.actorWatchCloseChan <- struct{}{}
	return nil
}


//check if actor key exist,otherwise add it
func (this *Registry) checkOrCreateActorRegistry(kv *mvccpb.KeyValue){
	id,_:=strconv.Atoi(regexp.MustCompile(`[/]actor[/]([0-9]+)`).ReplaceAllString(string(kv.Key),"$1"))
	actorId := util.ID(id)
	actorReg:=NewActorRegistry(actorId)
	json.Unmarshal(kv.Value,actorReg)
	log.Warnf("checkOrCreateActorRegistry",log.PrettyStruct(actorReg))
	this.GlobalRegistry.AddActor(actorReg)
}

func (this *Registry) deleteActorRegistry(kv *mvccpb.KeyValue){
	id,_:=strconv.Atoi(regexp.MustCompile(`[/]actor[/]([0-9]+)`).ReplaceAllString(string(kv.Key),"$1"))
	actorId := util.ID(id)
	this.GlobalRegistry.RemoveActor(actorId)
}

//check if process key exist,otherwise add it
func (this *Registry) checkOrCreateProcessRegistry(kv *mvccpb.KeyValue){
	id,_:=strconv.Atoi(regexp.MustCompile(`[/]processids[/]([0-9]+)`).ReplaceAllString(string(kv.Key),"$1"))
	pid := util.ProcessId(id)
	processReg:=NewProcessRegistry(pid)
	this.GlobalRegistry.AddProcess(processReg)
}

func (this *Registry) deleteProcessRegistry(kv *mvccpb.KeyValue) {
	id,_:=strconv.Atoi(regexp.MustCompile(`[/]processids[/]([0-9]+)`).ReplaceAllString(string(kv.Key),"$1"))
	pid := util.ProcessId(id)
	this.GlobalRegistry.RemoveProcess(pid)
}

//update actor registries via etcd
func (this *Registry) startUpdateRemoteActorInfo()error {
	log.Info("Registry:startUpdateRemoteActorInfo")
	client:=this.GetProcess().GetEtcd()
	res,err:=client.Get(context.TODO(),"/actor/",clientv3.WithPrefix())

	if err != nil {
		log.Error(err.Error())
		return err
	}
	for _,v:=range res.Kvs {
		this.checkOrCreateActorRegistry(v)
	}
	watcher:=client.Watch(context.TODO(),"/actor/",clientv3.WithPrefix())
	this.actorWatchCloseChan = make(chan struct{})
	go func() {
		for {
			select {
			case resp:=<-watcher:
				for _,e:=range resp.Events {
					if e.Type==mvccpb.PUT {
						log.Warnf("PUT actor",string(e.Kv.Key),string(e.Kv.Value))
						this.checkOrCreateActorRegistry(e.Kv)
					} else if e.Type == mvccpb.DELETE {
						log.Warnf("DELETE actor",string(e.Kv.Key),string(e.Kv.Value))
						this.deleteActorRegistry(e.Kv)
					}
				}
			case <-this.actorWatchCloseChan:
				goto CLOSE
			}
		}
	CLOSE:
		close(this.actorWatchCloseChan)
	}()
	return nil
}

//update process registries information via etcd
func (this *Registry) startUpdateRemoteProcessInfo()error{
	log.Info("Registry:startUpdateRemoteProcessInfo")
	client:=this.GetProcess().GetEtcd()
	res,err:=client.Get(context.TODO(),"/processids/",clientv3.WithPrefix())
	if err != nil {
		log.Error(err.Error())
		return err
	}
	for _,v:=range res.Kvs {
		this.checkOrCreateProcessRegistry(v)
	}
	watchChan:=client.Watch(context.TODO(),"/processids/",clientv3.WithPrefix(),clientv3.WithRev(res.Header.Revision))
	this.processWatchCloseChan = make(chan struct{})
	go func() {
		for {
			select {
			case resp:=<-watchChan:
				for _,e:=range resp.Events {
					if e.Type == mvccpb.PUT {
						log.Warnf("PUT Process Registry",string(e.Kv.Key))
						this.checkOrCreateProcessRegistry(e.Kv)
					} else if e.Type== mvccpb.DELETE {
						log.Warnf("DELETE Process Registry",string(e.Kv.Key))
						this.deleteProcessRegistry(e.Kv)
					}
				}
			case <-this.processWatchCloseChan:
				goto CLOSE
			}
		}
	CLOSE:
		close(this.processWatchCloseChan)
	}()
	return nil
}

func (this *Registry) RegisterProcessInfo()error {
	client:=this.GetProcess().GetEtcd()
	s,err:=json.Marshal(CreateProcessRegistryInfo(this.GetProcess()))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	res,err:=client.Put(context.TODO(),"/process/"+this.process.Id().String()+"/info",string(s))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	res,err=client.Put(context.TODO(),"/processids/"+this.process.Id().String()+"",time.Now().String())
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Warnf("res",res)
	return nil
}

func (this *Registry) RegisterActors()error{
	client:=this.GetProcess().GetEtcd()
	s,err:=json.Marshal(CreateProcessActorsInfo(this.GetProcess()))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	res,err:=client.Put(context.TODO(),"/process/"+this.process.Id().String()+"/actors",string(s))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Infof("res",res)
	return nil
}

func (this *Registry) RegistryServices(){

}

func (this *Registry) RegisterActorRemote(actor lokas.IActor)error {
	client:=this.GetProcess().GetEtcd()
	s,err:=json.Marshal(CreateActorRegistryInfo(actor))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	prefix := "/actor/"
	leaseId,isReg,err:=actor.GetLeaseId()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if !isReg {
		res,err:=client.Put(context.TODO(),prefix+actor.GetId().String(),string(s),clientv3.WithLease(leaseId))
		if err != nil {
			log.Error(err.Error())
			return err
		}
		log.Warnf("res",res)
	}
	return nil
}

func (this *Registry) UnregisterActorRemote(actor lokas.IActor)error{
	client:=this.GetProcess().GetEtcd()
	leaseId,_,err:=actor.GetLeaseId()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	_,err=client.Lease.Revoke(context.TODO(),leaseId)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Registry) RegisterServiceRemote(service *lokas.Service)error {
	panic("implement me")
}

func (this *Registry) UnregisterServiceRemote(service *lokas.Service)error {
	panic("implement me")
}

func (this *Registry) QueryRemoteActorsByType(typ string)[]*Actor{
	ret:=[]*Actor{}
	return ret
}

func (this *Registry) QueryRemoteActorsByServer(typ string,ServerId int32)[]*Actor{
	ret:=[]*Actor{}
	return ret
}

func (this *Registry) RegisterActorLocal(actor lokas.IActor)error {
	log.Infof("RegisterActorLocal",actor.GetId())
	re := &ActorRegistry{
		Id:        actor.GetId(),
		Type:      actor.Type(),
		ProcessId: this.process.Id(),
		GameId:    this.process.GameId(),
		Version:   this.process.Version(),
		ServerId:  this.process.ServerId(),
		Ts:        time.Now(),
	}
	this.LocalRegistry.AddActor(re)
	return nil
}

func (this *Registry) UnregisterActorLocal(actor lokas.IActor)error {
	this.LocalRegistry.RemoveActor(actor.GetId())
	return nil
}

//TODO
func (this *Registry) RegisterServiceLocal(service *lokas.Service)error {
	this.LocalRegistry.mu.Lock()
	defer this.LocalRegistry.mu.Unlock()
	se := this.LocalRegistry.Service[service.Id]
	if se == nil {
		se = &ServiceRegistry{
			Id:          0,
			ServiceType: 0,
			GameId:      "",
			Version:     "",
			ServerId:    0,
			Weights:     nil,
			Ts:          time.Time{},
		}
	}
	return nil
}

func (this *Registry) UnregisterServiceLocal(service *lokas.Service)error {
	this.LocalRegistry.mu.Lock()
	defer this.LocalRegistry.mu.Unlock()
	se := this.LocalRegistry.Service[service.Id]
	if se == nil {
		//TODO uni error
		log.Panic("cannot found service")
		return errors.New("cannont found service")
	}
	delete(se.Weights, service.ActorId)
	return nil
}
