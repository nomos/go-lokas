//this is a generate file,do not modify it!
package lox

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/ecs"
	"github.com/nomos/go-lokas/protocol"
	"reflect"
	"time"
)

var _ lokas.IComponent = (*AvatarInfo)(nil)

//角色信息
type AvatarInfo struct {
	ecs.Component `json:"-" bson:"-"`
	LastLogin time.Time 
	LastStaminaUpdate time.Time 
}

func (this *AvatarInfo) GetId()(protocol.BINARY_TAG,error){
	return protocol.GetTypeRegistry().GetTagByType(reflect.TypeOf(this).Elem())
}

func (this *AvatarInfo) Serializable()protocol.ISerializable {
	return this
}

func NewAvatarInfo()*AvatarInfo{
	ret:=&AvatarInfo{
		LastLogin: time.Now(),
		LastStaminaUpdate: time.Now(),
	}
	return ret
}

func (this *AvatarInfo) UpdateLogin(){
	this.LastLogin = time.Now()
}

func (this *AvatarInfo) OnAdd(e lokas.IEntity, r lokas.IRuntime) {

}

func (this *AvatarInfo) OnRemove(e lokas.IEntity, r lokas.IRuntime) {

}

func (this *AvatarInfo) OnCreate(r lokas.IRuntime) {

}

func (this *AvatarInfo) OnDestroy(r lokas.IRuntime) {

}