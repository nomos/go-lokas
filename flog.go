package lokas

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/log/flog"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"go.uber.org/zap"
)

func LogActorType(actor IActorInfo) zap.Field {
	return zap.String("actortype", actor.Type())
}

func LogActorId(actor IActorInfo) zap.Field {
	return zap.Int64("actorid", actor.GetId().Int64())
}

func LogActorProcessId(actor IActorInfo) zap.Field {
	return zap.Int32("pid", actor.PId().Int32())
}

func LogActorInfo(actor IActorInfo) log.ZapFields {
	ret := []zap.Field{}
	ret = append(ret, LogActorType(actor))
	ret = append(ret, LogActorId(actor))
	ret = append(ret, LogActorProcessId(actor))
	return ret
}

func LogServiceInfo(service *ServiceInfo) log.ZapFields {
	ret := []zap.Field{}
	ret = append(ret, flog.ServiceType(service.ServiceType))
	ret = append(ret, flog.LineId(service.LineId))
	ret = append(ret, flog.ServiceId(service.ServiceId))
	return ret
}

func LogModule(s IModule) zap.Field {
	return zap.String("module", s.Type())
}

func LogAvatarName(avatar IAvatarSession) zap.Field {
	return zap.String("username", avatar.GetUserName())
}

func LogAvatarUserId(avatar IAvatarSession) zap.Field {
	return zap.Int64("userid", avatar.GetUserId().Int64())
}

func LogAvatarSessionId(avatar IAvatarSession) zap.Field {
	return zap.Int64("avatarid", avatar.GetId().Int64())
}

func LogAvatarServer(avatar IAvatarSession) zap.Field {
	return zap.Int32("server", avatar.GetServerId())
}

func LogAvatarGate(avatar IAvatarSession) zap.Field {
	return zap.Int64("gateid", avatar.GetGateId().Int64())
}

func LogAvatarSessionInfo(avatar IAvatarSession) log.ZapFields {
	ret := log.ZapFields{}
	ret = ret.Append(LogAvatarName(avatar))
	ret = ret.Append(LogAvatarGate(avatar))
	ret = ret.Append(LogAvatarSessionId(avatar))
	ret = ret.Append(LogAvatarServer(avatar))
	ret = ret.Append(LogAvatarUserId(avatar))
	ret = ret.Append(flog.AvatarId(avatar.GetId()))
	return ret
}

func LogAvatarInfo(avatar IActor) log.ZapFields {
	a := avatar.(IAvatar)
	ret := log.ZapFields{}
	ret = ret.Concat(LogActorInfo(a))
	ret = ret.Append(LogAvatarName(a))
	ret = ret.Append(LogAvatarServer(a))
	ret = ret.Append(LogAvatarSessionId(a))
	ret = ret.Append(LogAvatarUserId(a))
	ret = ret.Append(flog.AvatarId(a.GetId()))
	return ret
}

func LogAvatarMsgInfo(avatar IAvatar, transId uint32, msg protocol.ISerializable, sess util.ID) log.ZapFields {
	return LogAvatarInfo(avatar).Append(flog.TransId(transId)).Concat(protocol.LogMsgInfo(msg)).Append(flog.ClientSessionId(sess))
}

func LogAvatarMsgDetail(avatar IAvatar, transId uint32, msg protocol.ISerializable, sess util.ID) log.ZapFields {
	return LogAvatarInfo(avatar).Append(flog.TransId(transId)).Concat(protocol.LogMsgDetail(msg)).Append(flog.ClientSessionId(sess))
}

func LogAvatarSessionMsgInfo(avatar IAvatarSession, transId uint32, msg protocol.ISerializable, sess util.ID) log.ZapFields {
	return LogAvatarSessionInfo(avatar).Append(flog.TransId(transId)).Concat(protocol.LogMsgInfo(msg)).Append(flog.ClientSessionId(sess))
}

func LogActorSendMsgInfo(actor IActorInfo, serializable protocol.ISerializable, transid uint32, toActorId util.ID) log.ZapFields {
	return LogActorInfo(actor).Concat(protocol.LogMsgInfo(serializable)).Append(flog.TransId(transid)).Append(flog.ToActorId(toActorId))
}

func LogActorReceiveMsgInfo(actor IActorInfo, serializable protocol.ISerializable, transid uint32, fromActorId util.ID) log.ZapFields {
	return LogActorInfo(actor).Concat(protocol.LogMsgInfo(serializable)).Append(flog.TransId(transid)).Append(flog.FromActorId(fromActorId))
}
