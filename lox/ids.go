package lox

import (
	"github.com/nomos/go-lokas/protocol"
	"reflect"
)

const (
	TAG_User = 130
	TAG_JWT =  131
	TAG_ClaimUser =  132
	TAG_AvatarMap =  133
	TAG_Avatar =  134
	TAG_AdminCmd = 135
	TAG_AdminCmdResult = 136
	TAG_CREATE_AVATAR = 137
	TAG_KICK_AVATAR = 138
	TAG_AVATAR_INFO = 139
	TAG_RESPONSE = 140
)

func init() {
	protocol.GetTypeRegistry().RegistryType(TAG_User,reflect.TypeOf((*User)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_JWT,reflect.TypeOf((*JwtClaim)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_RESPONSE,reflect.TypeOf((*Response)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_CREATE_AVATAR,reflect.TypeOf((*CreateAvatar)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_CREATE_AVATAR,reflect.TypeOf((*CreateAvatar)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_KICK_AVATAR,reflect.TypeOf((*KickAvatar)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_AVATAR_INFO,reflect.TypeOf((*AvatarInfo)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_ClaimUser,reflect.TypeOf((*ClaimUser)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_AvatarMap,reflect.TypeOf((*AvatarMap)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_Avatar,reflect.TypeOf((*Avatar)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_AdminCmd,reflect.TypeOf((*AdminCommand)(nil)).Elem())
	protocol.GetTypeRegistry().RegistryType(TAG_AdminCmdResult,reflect.TypeOf((*AdminCommandResult)(nil)).Elem())
}

