//this is a generate file,do not modify it!
package keys

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/ecs"
	"github.com/nomos/go-lokas/protocol"
	"reflect"
)

var _ lokas.IComponent = (*MouseEvent)(nil)

type MouseEvent struct {
	ecs.Component `json:"-" bson:"-"`
	Event MOUSE_EVENT_TYPE 
	Button MOUSE_BUTTON 
	X int32 
	Y int32 
	Num int64 
}

func (this *MouseEvent) GetId()(protocol.BINARY_TAG,error){
	return protocol.GetTypeRegistry().GetTagByType(reflect.TypeOf(this).Elem())
}

func (this *MouseEvent) Serializable()protocol.ISerializable {
	return this
}
