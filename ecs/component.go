package ecs

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/protocol"
	"reflect"
)

type Component struct {
	dirty   bool
	runtime lokas.IRuntime
	entity  lokas.IEntity
}

func (this *Component) SetDirty(d bool) {
	this.dirty = d
	if this.entity!=nil&&d==true {
		this.entity.SetDirty(true)
	}
}

func (this *Component) GetEntity()lokas.IEntity {
	return this.entity
}

func (this *Component) SetEntity(e lokas.IEntity) {
	this.entity = e
	this.dirty = true
}

func (this *Component) SetRuntime(r lokas.IRuntime) {
	this.runtime = r
	this.dirty = true
}

func (this *Component) GetRuntime()lokas.IRuntime {
	return this.runtime
}

func (this *Component) GetComponentName()string{
	return protocol.GetTypeRegistry().GetNameByType(reflect.TypeOf(this))
}

func (this *Component) GetSibling(t protocol.BINARY_TAG) lokas.IComponent {
	return this.entity.Get(t)
}
