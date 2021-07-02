package logfield

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"go.uber.org/zap"
	"reflect"
)

func ActorType(actor lokas.IActorInfo)zap.Field{
	return zap.String("actor_type",actor.Type())
}

func ActorId(actor lokas.IActorInfo)zap.Field{
	return zap.Int64("actor_id",actor.GetId().Int64())
}

func ProcessId(actor lokas.IActorInfo)zap.Field{
	return zap.Int32("pid",actor.PId().Int32())
}

func ActorInfo(actor lokas.IActorInfo)log.ZapFields{
	ret:=[]zap.Field{}
	ret = append(ret, ActorType(actor))
	ret = append(ret, ActorId(actor))
	ret = append(ret, ProcessId(actor))
	return ret
}

func FuncClass(s interface{})zap.Field{
	return zap.String("class",reflect.TypeOf(s).Elem().Name())
}

func FuncName(s string)zap.Field{
	return zap.String("func",s)
}
func Result(s interface{})zap.Field{
	return zap.Any("result",s)
}

func Key(key interface{})zap.Field{
	return zap.Any("key",key)
}

func Value(value interface{})zap.Field{
	return zap.Any("value",value)
}

func KeyValue(key interface{},value interface{})log.ZapFields{
	ret:=[]zap.Field{}
	ret = append(ret, Key(key))
	ret = append(ret, Value(value))
	return ret
}

func FuncInfo(class interface{},f string)log.ZapFields{
	ret:=[]zap.Field{}
	ret = append(ret, FuncClass(class))
	ret = append(ret, FuncName(f))
	return ret
}

func MsgName(serializable protocol.ISerializable)zap.Field{
	id,_:=serializable.GetId()
	return zap.String("msg_name",id.String())
}

func MsgId(serializable protocol.ISerializable)zap.Field{
	id,_:=serializable.GetId()
	return zap.Int32("msg_id",int32(id))
}

func MsgInfo(serializable protocol.ISerializable)log.ZapFields{
	ret:=[]zap.Field{}
	ret = append(ret, MsgName(serializable))
	ret = append(ret, MsgId(serializable))
	return ret
}

func ErrorMessage(err protocol.IError)zap.Field {
	return zap.String("err_msg",err.Error())
}

func ErrorCode(err protocol.IError)zap.Field {
	return zap.Int("err_code",err.ErrCode())
}

func ErrorInfo(err protocol.IError)[]zap.Field{
	ret:=[]zap.Field{}
	ret = append(ret, ErrorMessage(err))
	ret = append(ret, ErrorCode(err))
	return ret
}

func Error(err error)zap.Field{
	return zap.Error(err)
}

func Offset(offset int)zap.Field {
	return zap.Int("offset",offset)
}

func Size(size int)zap.Field {
	return zap.Int("size",size)
}

func Size64(size int64)zap.Field {
	return zap.Int64("size",size)
}

func Module(s lokas.IModule)zap.Field{
	return zap.String("module",s.Type())
}

func Type(t string)zap.Field {
	return zap.String("type",t)
}

func Path(p string)zap.Field {
	return zap.String("path",p)
}

func Address(addr string)zap.Field {
	return zap.String("address",addr)
}

func DataBase(d string)zap.Field {
	return zap.String("database",d)
}

func Success(success bool)zap.Field{
	return zap.Bool("success",success)
}

func TransId(transid uint32 )zap.Field{
	return zap.Uint32("trans_id",transid)
}

func ToActorId(id util.ID )zap.Field{
	return zap.Int64("to_actor_id",id.Int64())
}

func FromActorId(id util.ID )zap.Field{
	return zap.Int64("from_actor_id",id.Int64())
}

func ActorSendMsgInfo(actor lokas.IActorInfo,serializable protocol.ISerializable,transid uint32,toActorId util.ID)log.ZapFields{
	return ActorInfo(actor).Concat(MsgInfo(serializable)).Append(TransId(transid)).Append(ToActorId(toActorId))
}

func ActorReceiveMsgInfo(actor lokas.IActorInfo,serializable protocol.ISerializable,transid uint32,fromActorId util.ID)log.ZapFields{
	return ActorInfo(actor).Concat(MsgInfo(serializable)).Append(TransId(transid)).Append(FromActorId(fromActorId))
}

func ActorRouterMsgInfo(serializable protocol.ISerializable,transid uint32,fromActorId util.ID,toActorId util.ID)log.ZapFields{
	return MsgInfo(serializable).Append(TransId(transid)).Append(FromActorId(fromActorId)).Append(ToActorId(toActorId))
}