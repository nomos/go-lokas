package lox

import (
	"github.com/nomos/go-lokas/protocol"
	"reflect"
)

const (
	TAG_USER             = 130
	TAG_JWT              = 131
	TAG_CLAIM_USER       = 132
	TAG_AVATAR_MAP       = 133
	TAG_AVATAR           = 134
	TAG_ADMIN_CMD        = 135
	TAG_ADMIN_CMD_RESULT = 136
	TAG_CREATE_AVATAR    = 137
	TAG_KICK_AVATAR      = 138
	TAG_RESPONSE         = 140
	TAG_CONSOLE_EVENT    = 221
)

func init() {
	protocol.GetTypeRegistry().RegistryType(TAG_USER, reflect.TypeOf((*User)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_JWT, reflect.TypeOf((*JwtClaim)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_RESPONSE, reflect.TypeOf((*Response)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_CREATE_AVATAR, reflect.TypeOf((*CreateAvatar)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_KICK_AVATAR, reflect.TypeOf((*KickAvatar)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_CLAIM_USER, reflect.TypeOf((*ClaimUser)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_AVATAR_MAP, reflect.TypeOf((*AvatarMap)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_AVATAR, reflect.TypeOf((*Avatar)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_ADMIN_CMD, reflect.TypeOf((*AdminCommand)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_ADMIN_CMD_RESULT, reflect.TypeOf((*AdminCommandResult)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_CONSOLE_EVENT, reflect.TypeOf((*ConsoleEvent)(nil)).Elem())
}
