package lox

import (
	"sync"
	"time"

	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
)

func NewCommonRegistry() *CommonRegistry {
	ret := &CommonRegistry{
		Processes:      map[util.ProcessId]*ProcessRegistry{},
		Service:        map[lokas.ServiceType]map[uint16]*ServiceRegistry{},
		Actors:         map[util.ID]*ActorRegistry{},
		ActorsByType:   map[string][]util.ID{},
		ActorsByServer: map[int32][]util.ID{},
		Ts:             time.Time{},
	}
	return ret
}

type CommonRegistry struct {
	Processes      map[util.ProcessId]*ProcessRegistry
	Service        map[lokas.ServiceType]map[uint16]*ServiceRegistry
	Actors         map[util.ID]*ActorRegistry
	ActorsByType   map[string][]util.ID
	ActorsByServer map[int32][]util.ID
	Ts             time.Time
	mu             sync.Mutex
}

func (this *CommonRegistry) GetActorRegistry(id util.ID) *ActorRegistry {
	return this.Actors[id]
}

func (this *CommonRegistry) GetActorIdsByTypeAndServerId(serverId int32, typ string) []util.ID {
	ret := []util.ID{}
	serverIds, ok := this.ActorsByServer[serverId]
	if !ok {
		return ret
	}
	typeIds, ok := this.ActorsByType[typ]
	if !ok {
		return ret
	}
	for _, a := range serverIds {
		for _, b := range typeIds {
			if a == b {
				ret = append(ret, a)
			}
		}
	}
	return ret
}

func removeIdFromArr(id util.ID, arr []util.ID) []util.ID {
	ret := []util.ID{}
	for _, v := range arr {
		if v != id {
			ret = append(ret, id)
		}
	}
	return ret
}

func addId2ArrOnce(id util.ID, arr []util.ID) []util.ID {
	for _, v := range arr {
		if v == id {
			return arr
		}
	}
	arr = append(arr, id)
	return arr
}

func (this *CommonRegistry) AddActor(actor *ActorRegistry) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.Actors[actor.Id] = actor
	if actorArr, ok := this.ActorsByType[actor.Type]; ok {
		actorArr = addId2ArrOnce(actor.Id, actorArr)
		this.ActorsByType[actor.Type] = actorArr
	} else {
		this.ActorsByType[actor.Type] = []util.ID{actor.Id}
	}
	if actorArr, ok := this.ActorsByServer[actor.ServerId]; ok {
		actorArr = addId2ArrOnce(actor.Id, actorArr)
		this.ActorsByServer[actor.ServerId] = actorArr
	} else {
		this.ActorsByServer[actor.ServerId] = []util.ID{actor.Id}
	}
}

func (this *CommonRegistry) RemoveActor(actorId util.ID) {
	this.mu.Lock()
	defer this.mu.Unlock()
	if actor, ok := this.Actors[actorId]; ok {
		delete(this.Actors, actorId)
		if actorArr, ok := this.ActorsByType[actor.Type]; ok {
			this.ActorsByType[actor.Type] = removeIdFromArr(actor.Id, actorArr)
		}
		if actorArr, ok := this.ActorsByServer[actor.ServerId]; ok {
			this.ActorsByServer[actor.ServerId] = removeIdFromArr(actor.Id, actorArr)
		}
	}
}

func (this *CommonRegistry) AddProcess(process *ProcessRegistry) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.Processes[process.Id] = process
}

func (this *CommonRegistry) RemoveProcess(id util.ProcessId) {
	this.mu.Lock()
	defer this.mu.Unlock()
	delete(this.Processes, id)
}

func (this *CommonRegistry) AddService(service *ServiceRegistry) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.Service[service.ServiceType][service.ServiceId] = service
}

// func (this *CommonRegistry) RemoveService(id protocol.BINARY_TAG) {
// 	this.mu.Lock()
// 	defer this.mu.Unlock()
// 	delete(this.Service, id)
// }

func (this *CommonRegistry) RemoveService(serviceType lokas.ServiceType, serviceId uint16) {
	this.mu.Lock()
	defer this.mu.Unlock()
	delete(this.Service[serviceType], serviceId)
}

// service
type ServiceRegistry struct {
	Id          uint32
	ServiceType lokas.ServiceType
	ServiceId   uint16
	// GameId      string
	Host     string
	Port     uint32
	Version  string
	Cnt      uint32
	CreateAt time.Time
	// Weights map[util.ID]int
	// Ts time.Time
}

func NewServiceRegistry(serviceType lokas.ServiceType, serviceId uint16) *ServiceRegistry {
	return &ServiceRegistry{
		Id:          uint32(serviceId) * 100,
		ServiceType: serviceType,
		ServiceId:   serviceId,
		Host:        "",
		Port:        0,
		Version:     "",
		Cnt:         0,
		CreateAt:    time.Time{},
	}
}

type ActorRegistry struct {
	Id        util.ID
	ProcessId util.ProcessId
	Type      string
	GameId    string
	Version   string
	ServerId  int32
	//Health    lokas.ActorState
	Ts time.Time
}

func NewActorRegistry(id util.ID) *ActorRegistry {
	ret := &ActorRegistry{
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

func NewProcessRegistry(id util.ProcessId) *ProcessRegistry {
	ret := &ProcessRegistry{
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
	Id     util.ProcessId
	Actors []util.ID
	Ts     time.Time
}

func CreateProcessActorsInfo(process lokas.IProcess) *ProcessActorsInfo {
	ret := &ProcessActorsInfo{
		Id:     process.PId(),
		Actors: process.GetActorIds(),
		Ts:     time.Now(),
	}
	return ret
}

type ActorRegistryInfo struct {
	Id        util.ID
	Type      string
	ProcessId util.ProcessId
	GameId    string
	Version   string
	ServerId  int32
	Ts        time.Time
}

func CreateActorRegistryInfo(actor lokas.IActor) *ActorRegistryInfo {
	ret := &ActorRegistryInfo{
		Id:        actor.GetId(),
		Type:      actor.Type(),
		ProcessId: actor.GetProcess().PId(),
		GameId:    actor.GetProcess().GameId(),
		Version:   actor.GetProcess().Version(),
		ServerId:  actor.GetProcess().ServerId(),
		Ts:        time.Now(),
	}
	return ret
}

type ProcessServiceInfo struct {
	Id       util.ProcessId
	Services map[protocol.BINARY_TAG]int
}

type ProcessRegistryInfo struct {
	Id       util.ProcessId
	GameId   string
	Version  string
	ServerId int32
	Host     string
	Port     string
	Ts       time.Time
}

func CreateProcessRegistryInfo(process lokas.IProcess) *ProcessRegistryInfo {
	ret := &ProcessRegistryInfo{
		Id:       process.PId(),
		GameId:   process.GameId(),
		Version:  process.Version(),
		ServerId: process.ServerId(),
		Host:     process.Config().GetString("host"),
		Port:     process.Config().GetString("port"),
		Ts:       time.Now(),
	}
	return ret
}
