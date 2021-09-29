package flog

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"go.uber.org/zap"
)

func AvatarName(avatar lokas.IAvatarSession) zap.Field {
	return zap.String("username", avatar.GetUserName())
}

func AvatarSessionId(avatar lokas.IAvatarSession) zap.Field {
	return zap.Int64("avatarid", avatar.GetId().Int64())
}

func AvatarServer(avatar lokas.IAvatarSession) zap.Field {
	return zap.Int32("server", avatar.GetServerId())
}

func AvatarGate(avatar lokas.IAvatarSession) zap.Field {
	return zap.Int64("gateid", avatar.GetGateId().Int64())
}

func AvatarInfo(avatar lokas.IAvatar) log.ZapFields {
	ret := log.ZapFields{}
	ret = ret.Concat(ActorInfo(avatar))
	ret = ret.Append(AvatarName(avatar))
	ret = ret.Append(AvatarServer(avatar))
	ret = ret.Append(AvatarSessionId(avatar))
	return ret
}

func ClientSessionId(id util.ID) zap.Field {
	return zap.Int64("clientsession", id.Int64())
}

func AvatarMsgInfo(avatar lokas.IAvatar, transId uint32, msg protocol.ISerializable, sess util.ID) log.ZapFields {
	return AvatarInfo(avatar).Append(TransId(transId)).Concat(MsgInfo(msg)).Append(ClientSessionId(sess))
}

func UserName(name string)zap.Field{
	return zap.String("username", name)
}

func AvatarId(id util.ID)zap.Field{
	return zap.Int64("avatarid", id.Int64())
}

func ProcessId(id util.ID)zap.Field{
	return zap.Int64("processid", id.Int64())
}

func ServerId(id int32)zap.Field{
	return zap.Int32("serverid", id)
}

func GameId(name string)zap.Field{
	return zap.String("gameid", name)
}

func RegElapsedDay(day int)zap.Field{
	return zap.Int("rentday", day)
}

func AvatarSessionInfo(avatar lokas.IAvatarSession) log.ZapFields {
	ret := log.ZapFields{}
	ret = ret.Append(AvatarName(avatar))
	ret = ret.Append(AvatarGate(avatar))
	ret = ret.Append(AvatarSessionId(avatar))
	ret = ret.Append(AvatarServer(avatar))
	return ret
}

func AvatarSessionMsgInfo(avatar lokas.IAvatarSession, transId uint32, msg protocol.ISerializable, sess util.ID) log.ZapFields {
	return AvatarSessionInfo(avatar).Append(TransId(transId)).Concat(MsgInfo(msg)).Append(ClientSessionId(sess))
}
