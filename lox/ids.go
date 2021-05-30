package lox

import (
	"github.com/nomos/go-lokas/protocol"
	"reflect"
)

const (
	TAG_User = 70
	TAG_JWT =  71
)

func init() {
	protocol.GetTypeRegistry().RegistryType(TAG_User,reflect.TypeOf((*User)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_JWT,reflect.TypeOf((*JwtClaim)(nil)).Elem())
}

