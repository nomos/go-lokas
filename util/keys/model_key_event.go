//this is a generate file,do not modify it!
package keys

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/ecs"
	"github.com/nomos/go-lokas/protocol"
	"reflect"
)

var _ lokas.IComponent = (*KeyEvent)(nil)

type KeyEvent struct {
	ecs.Component `json:"-" bson:"-"`
	Code KEY 
	Event KEY_EVENT_TYPE 
}

func (this *KeyEvent) GetId()(protocol.BINARY_TAG,error){
	return protocol.GetTypeRegistry().GetTagByType(reflect.TypeOf(this).Elem())
}

func (this *KeyEvent) Serializable()protocol.ISerializable {
	return this
}
