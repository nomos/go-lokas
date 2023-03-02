package flog

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/util"
	"go.uber.org/zap"
	"reflect"
)

func FuncClass(s interface{}) zap.Field {
	return zap.String("class", reflect.TypeOf(s).Elem().Name())
}

func FuncName(s string) zap.Field {
	return zap.String("func", s)
}
func Result(s interface{}) zap.Field {
	return zap.Any("result", s)
}

func FuncInfo(class interface{}, f string) log.ZapFields {
	ret := []zap.Field{}
	ret = append(ret, FuncClass(class))
	ret = append(ret, FuncName(f))
	return ret
}

func Error(err error) zap.Field {
	return zap.Error(err)
}

func Offset(offset int) zap.Field {
	return zap.Int("offset", offset)
}

func Size(size int) zap.Field {
	return zap.Int("size", size)
}

func Size64(size int64) zap.Field {
	return zap.Int64("size", size)
}

func Type(t string) zap.Field {
	return zap.String("type", t)
}

func Path(p string) zap.Field {
	return zap.String("path", p)
}

func Address(addr string) zap.Field {
	return zap.String("address", addr)
}

func DataBase(d string) zap.Field {
	return zap.String("database", d)
}

func Success(success bool) zap.Field {
	return zap.Bool("success", success)
}

func TransId(transid uint32) zap.Field {
	return zap.Uint32("trans_id", transid)
}

func ToActorId(id util.ID) zap.Field {
	return zap.Int64("to_actor_id", id.Int64())
}

func ReqMsg(req bool) zap.Field {
	return zap.Bool("is_req", req)
}

func FromActorId(id util.ID) zap.Field {
	return zap.Int64("from_actor_id", id.Int64())
}

func ToPid(id util.ProcessId) zap.Field {
	return zap.Int64("to_pid", id.Int64())
}
