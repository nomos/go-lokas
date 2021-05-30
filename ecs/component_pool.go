package ecs

import (
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas"
	"reflect"
)


type ComponentCreator func(args...interface{}) lokas.IComponent

var _ lokas.IComponentPool = (*ComponentPool)(nil)

type ComponentPool struct {
	creator       ComponentCreator
	componentType reflect.Type
	pool          []lokas.IComponent
	admin         lokas.IRuntime
	itemCount     int
}

func NewComponentPool(creator ComponentCreator,comp lokas.IComponent) lokas.IComponentPool {
	ret:= &ComponentPool{
		creator:       creator,
		componentType: reflect.TypeOf(comp),
		pool: make([]lokas.IComponent,0),
	}
	return ret
}



func (this *ComponentPool) Get()lokas.IComponent{
	length:=len(this.pool)
	if length==0 {
		return this.Create()
	}
	return this.Pop()
}

func (this *ComponentPool) Recycle(comp lokas.IComponent){
	if reflect.TypeOf(comp)==this.componentType {
		this.pool = append(this.pool, comp)
	} else {
		log.Error("comp type mismatch!")
	}
}

func (this *ComponentPool) Create(args...interface{}) lokas.IComponent {
	ret := this.creator(args...)
	ret.SetRuntime(this.admin)
	ret.SetDirty(true)
	ret.OnCreate(this.admin)
	this.itemCount++
	return ret;
}

func (this *ComponentPool) PopAndDestroy(){
	length:=len(this.pool)
	comp:=this.pool[length-1]
	comp.OnDestroy(this.admin)
	this.pool = this.pool[:length-1]
	this.itemCount--
}

func (this *ComponentPool) Destroy(){

}

func (this *ComponentPool) Pop() lokas.IComponent {
	length:=len(this.pool)
	if length==0 {
		return nil
	}
	comp:=this.pool[length-1]
	this.pool = this.pool[:length-1]
	this.itemCount--
	return comp
}