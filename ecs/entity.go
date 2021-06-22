package ecs

import (
	"github.com/nomos/go-events"
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"reflect"
)

type Entity struct {
	id            util.ID
	dirty         bool
	onDestroy     bool
	runtime       lokas.IRuntime
	eventListener events.EventEmmiter
	step          int64
	removeMarks   []protocol.BINARY_TAG
	modifyMarks   []protocol.BINARY_TAG
	addMarks      []protocol.BINARY_TAG
	components    map[protocol.BINARY_TAG]lokas.IComponent
}

func (this *Entity) Components()map[protocol.BINARY_TAG]lokas.IComponent {
	return this.components
}

func (this *Entity) Clone()*Entity{
	panic("impl")
}

func (this *Entity) SetId(id util.ID) {
	this.id = id
}

func (this *Entity) GetId()util.ID {
	return this.id
}

func (this *Entity) Init(){

}

func CreateEntity() lokas.IEntity {
	ret := &Entity{
		removeMarks:[]protocol.BINARY_TAG{},
		modifyMarks:[]protocol.BINARY_TAG{},
		addMarks:[]protocol.BINARY_TAG{},
		components:map[protocol.BINARY_TAG]lokas.IComponent{},
	}
	ret.Init()
	return ret
}

func (this *Entity) Add(c lokas.IComponent) {
	if util.IsNil(c) {
		return
	}
	id,err:=c.GetId()
	if err != nil {
		log.Panic(err.Error())
	}
	if this.components[id]!=nil {
		return
	}
	c.SetEntity(this)
	c.SetRuntime(this.runtime)
	this.components[id] = c
	c.OnAdd(this,this.runtime)
	this.SetDirty(true)
}

func (this *Entity) AddByTag(t protocol.BINARY_TAG)lokas.IComponent {
	return nil
}

func (this *Entity) Remove(t protocol.BINARY_TAG)lokas.IComponent {
	comp:=this.components[t]
	if comp!=nil {
		comp.OnRemove(this,this.runtime)
	}
	delete(this.components,t)
	return comp
}

func (this *Entity) RemoveAll(){
	for _,v:=range this.components {
		if v!=nil {
			v.OnRemove(this,this.runtime)
		}
	}
	this.components = map[protocol.BINARY_TAG]lokas.IComponent{}
}

func (this *Entity) Get(t protocol.BINARY_TAG)lokas.IComponent {
	return this.components[t]
}

func (this *Entity) GetOrCreate(t protocol.BINARY_TAG)lokas.IComponent {
	c,ok:=this.components[t]
	if ok {
		return c
	}
	a,_ := protocol.GetTypeRegistry().GetInterfaceByTag(t)
	c=a.(lokas.IComponent)
	c.OnCreate(this.runtime)
	this.Add(c)
	return c
}

func (this *Entity) cleanup() {
	this.dirty = false
	this.removeMarks = []protocol.BINARY_TAG{}
	this.addMarks = []protocol.BINARY_TAG{}
	this.modifyMarks = []protocol.BINARY_TAG{}
}

func (this *Entity) Dirty()bool {
	return this.dirty
}

func (this *Entity) SetDirty(d bool) {
	this.dirty = d
	if d==false {
		for _,c:=range this.components {
			c.SetDirty(false)
		}
	}
}

func (this *Entity) hasTypeInComponents(t reflect.Type)bool {
	for _,comp:=range this.components {
		if reflect.TypeOf(comp) == t {
			return true
		}
	}
	return false
}

func (this *Entity)Includes(componentTypeArray []reflect.Type)bool {
	for _,typ:=range componentTypeArray {
		if !this.hasTypeInComponents(typ) {
			return false
		}
	}
	return true
}

func (this *Entity) markModify(typ reflect.Type) {

}

func (this *Entity) MarkDirty(c lokas.IComponent) {
	if this.dirty {
		return
	}
	this.runtime.MarkDirtyEntity(this)
	this.dirty = true
	if this.runtime.IsServer() {
		this.step = this.runtime.CurrentTick()
	}
	if !GetComponentSyncAble(this.runtime,c) {
		return
	}
	this.markModify(reflect.TypeOf(c))
}

func (this *Entity) MarkDirtyByName(name string) {
	this.Dirty()

	if !GetComponentSyncAble(this.runtime,name) {
		return
	}
	typ := this.runtime.GetComponentType(name)
	this.markModify(typ)
}
