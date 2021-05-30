package ecs

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"reflect"
)

type EntityRef struct {
	runtime lokas.IRuntime
	Id util.ID
}

func (this *EntityRef) Entity()lokas.IEntity{
	return this.runtime.GetEntity(this.Id)
}

func init(){
	protocol.GetTypeRegistry().RegistryType(protocol.TAG_EntityRef,reflect.TypeOf((*EntityRef)(nil)).Elem())
}