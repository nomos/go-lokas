package lox

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/log/logfield"
	"github.com/nomos/go-lokas/util"
	"go.uber.org/zap"
	"sync"
)

var _ lokas.IActorContainer = (*ActorContainer)(nil)
var _ lokas.IModule = (*ActorContainer)(nil)

type ActorContainer struct {
	process lokas.IProcess
	Actors map[util.ID]lokas.IActor
	mu sync.Mutex
}

func (this *ActorContainer) Type() string {
	return "ActorContainer"
}

func (this *ActorContainer) Load(conf lokas.IConfig) error {
	return nil
}

func (this *ActorContainer) Unload() error {
	return nil
}

func (this *ActorContainer) GetProcess() lokas.IProcess {
	return this.process
}

func (this *ActorContainer) SetProcess(process lokas.IProcess) {
	this.process = process
}

func (this *ActorContainer) Start() error {
	return nil
}

func (this *ActorContainer) Stop() error {
	return nil
}

func (this *ActorContainer) OnStart() error {
	return nil
}

func (this *ActorContainer) OnStop() error {
	return nil
}

func NewActorContainer(process lokas.IProcess) *ActorContainer {
	ret := &ActorContainer{
		process: process,
		Actors: make(map[util.ID]lokas.IActor),
	}
	return ret
}

func (this *ActorContainer) GetActor(id util.ID)lokas.IActor{
	return this.Actors[id]
}

func (this *ActorContainer) StartActor(actor lokas.IActor){
	log.Info("starting",zap.String("module",actor.Type()))
	go func() {
		err:=actor.Start()
		if err != nil {
			log.Error("ActorContainer:StartActor:error",logfield.ActorInfo(actor).Append(logfield.Error(err))...)
			this.RemoveActor(actor)
		} else {
			log.Info("ActorContainer:StartActor:success",logfield.ActorInfo(actor)...)
			actor.OnStart()
		}
	}()
}

func (this *ActorContainer) StopActor(actor lokas.IActor){
	log.Info("stoping",zap.String("module",actor.Type()))
	go func() {
		err:=actor.Stop()
		if err != nil {
			log.Error("Actor stop error type:"+actor.Type()+" Id:"+actor.GetId().String()+" err:"+err.Error())
		} else {
			log.Info("stop success",zap.String("module",actor.Type()))
			actor.OnStop()
		}
	}()
}

func (this *ActorContainer) AddActor(actor lokas.IActor){
	this.mu.Lock()
	defer this.mu.Unlock()
	actor.SetProcess(this.process)
	this.Actors[actor.GetId()] = actor
	this.StartActor(actor)
	this.process.RegisterActors()
}

func (this *ActorContainer) RemoveActor(actor lokas.IActor){
	this.mu.Lock()
	defer this.mu.Unlock()
	this.RemoveActorById(actor.GetId())
	this.process.RegisterActors()
}

func (this *ActorContainer) RemoveActorById(id util.ID) lokas.IActor {
	actor,ok:=this.Actors[id]
	if ok {
		delete(this.Actors,id)
		this.process.RegisterActors()
		this.StopActor(actor)
		return actor
	}
	return nil
}

func (this *ActorContainer) GetActorIds()[]util.ID{
	ret:=[]util.ID{}
	for k,_:=range this.Actors {
		ret = append(ret, k)
	}
	return ret
}

