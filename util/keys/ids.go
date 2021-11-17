//this is a generated file,do not modify it!!!
package keys

import (
	"github.com/nomos/go-lokas/protocol"
	"reflect"
)

const (
	TAG_KEY_EVENT  protocol.BINARY_TAG = 311
	TAG_MOUSE_EVENT  protocol.BINARY_TAG = 312
)

func init() {
	protocol.GetTypeRegistry().RegistryType(TAG_KEY_EVENT,reflect.TypeOf((*KeyEvent)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_MOUSE_EVENT,reflect.TypeOf((*MouseEvent)(nil)).Elem())
}


