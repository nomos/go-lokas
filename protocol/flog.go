package protocol

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/log/flog"
	"github.com/nomos/go-lokas/util"
	"go.uber.org/zap"
)

func LogActorRouterMsgInfo(cmdId BINARY_TAG, transid uint32, fromActorId util.ID, toActorId util.ID, toPid util.ProcessId, isReq bool) log.ZapFields {
	ret := log.ZapFields{}
	ret.Append(LogCmdId(cmdId))
	ret.Append(flog.TransId(transid))
	ret.Append(flog.FromActorId(fromActorId))
	ret.Append(flog.ToActorId(toActorId))
	ret.Append(flog.ReqMsg(isReq))
	ret.Append(flog.ToPid(toPid))
	return ret
}

func LogCmdId(cmd BINARY_TAG) zap.Field {
	return zap.Uint16("cmdid", uint16(cmd))
}

func LogMsgName(serializable ISerializable) zap.Field {
	id, _ := serializable.GetId()
	return zap.String("msg_name", id.String())
}

func LogMsgId(serializable ISerializable) zap.Field {
	id, _ := serializable.GetId()
	return zap.Int32("msg_id", int32(id))
}

func LogMsgBody(serializable ISerializable) zap.Field {
	return zap.String("msg_body", log.PrettyStruct(serializable))
}

func LogMsgDetail(serializable ISerializable) log.ZapFields {
	ret := []zap.Field{}
	ret = append(ret, LogMsgName(serializable))
	ret = append(ret, LogMsgId(serializable))
	ret = append(ret, LogMsgBody(serializable))
	return ret
}

func LogMsgInfo(serializable ISerializable) log.ZapFields {
	ret := []zap.Field{}
	ret = append(ret, LogMsgName(serializable))
	ret = append(ret, LogMsgId(serializable))
	//ret = append(ret, MsgBody(serializable))
	return ret
}

func LogErrorMessage(err IError) zap.Field {
	return zap.String("err_msg", err.Error())
}

func LogErrorCode(err IError) zap.Field {
	return zap.Int("err_code", err.ErrCode())
}

func LogErrorInfo(err IError) []zap.Field {
	ret := []zap.Field{}
	ret = append(ret, LogErrorMessage(err))
	ret = append(ret, LogErrorCode(err))
	return ret
}
