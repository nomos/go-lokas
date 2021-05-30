package ecs

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/util"
	"reflect"
)


var _ lokas.IRuntime = (*Runtime)(nil)

type Runtime struct {
	timer *util.Timer
	sign chan<- int
	objContainer map[string]interface{}
	entityPool map[util.ID]lokas.IEntity
 	isServer bool
	tick int64
}

func CreateECS(updateTime int64,timeScale float32,server bool) lokas.IRuntime {
	var ret lokas.IRuntime =&Runtime{}
	ret.Init(updateTime,timeScale,server)
	return ret
}

func (this *Runtime) Init(updateTime int64,timeScale float32,server bool) {
	this.sign = make(chan int ,0)
	this.timer = util.CreateTimer(updateTime,timeScale,this.sign)
}

func (this *Runtime) GetEntity(id util.ID)lokas.IEntity {
	return this.entityPool[id]
}

func (this *Runtime) GetContext(name string)interface{} {
	return this.objContainer[name]
}

func (this *Runtime) SetContext(name string,value interface{}) {
	this.objContainer[name] = value
}

func (this *Runtime) CurrentTick()int64 {
	return this.tick
}

func (this *Runtime) Start() {

}

func (this *Runtime) Stop() {

}

func (this *Runtime) RunningTime()int64 {
	return 0
}

func (this *Runtime) GetTimeScale()float32 {
	return 0.
}

func (this *Runtime) SetTimeScale(scale float32){

}

func (this *Runtime) RegisterComponent(name string,c lokas.IComponent) {

}

func (this *Runtime) RegisterSingleton(name string,c lokas.IComponent) {

}

func (this *Runtime) GetComponentType(name string)reflect.Type {
	return reflect.TypeOf(nil)
}

func (this *Runtime) IsSyncAble(compName string) bool {
	return false
}

func (this *Runtime) CreateEntity()lokas.IEntity {
	return CreateEntity()
}

func (this *Runtime) IsServer() bool {
	return this.isServer
}

func (this *Runtime) MarkDirtyEntity(e lokas.IEntity){

}



