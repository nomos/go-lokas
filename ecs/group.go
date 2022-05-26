package ecs

import (
	"github.com/nomos/go-lokas/util"
	"github.com/nomos/go-lokas/util/slice"
	"reflect"
)

type Group interface {
	Init(compGroup []string, ecs Runtime)
	AddEntity(e Entity)
}

type group struct {
	hash           []string
	ecs            Runtime
	componentTypes []reflect.Type
	componentNames []string
	entityIndexes  []util.ID
	dirtyEntities  []Entity
}

func CreateGroup(compGroup []string, ecs Runtime) Group {
	var ret Group = &group{}
	ret.Init(compGroup, ecs)
	return ret
}

func (this *group) Init(compGroup []string, ecs Runtime) {
	for _, comp := range compGroup {
		compType := ecs.GetComponentType(comp)

		if slice.HasInterface(this.componentTypes, compType) {
			panic("Already has same type")
		}
		this.componentTypes = append(this.componentTypes, compType)
		this.componentNames = append(this.componentNames, comp)
	}
	this.ecs = ecs
}

func (this *group) AddEntity(e Entity) {
	if e.Includes(this.componentTypes) {

	}
}

func (this *group) HasEntity(e Entity) bool {
	return slice.Has(this.entityIndexes, e.id)
}
