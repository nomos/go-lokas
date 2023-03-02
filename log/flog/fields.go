package flog

import (
	"github.com/nomos/go-lokas/util"
	"go.uber.org/zap"
)

func ClientSessionId(id util.ID) zap.Field {
	return zap.Int64("clientsession", id.Int64())
}

func UserName(name string) zap.Field {
	return zap.String("username", name)
}

func AvatarId(id util.ID) zap.Field {
	return zap.Int64("avatarid", id.Int64())
}

func UserId(id util.ID) zap.Field {
	return zap.Int64("userid", id.Int64())
}

func ProcessId(id util.ID) zap.Field {
	return zap.Int64("processid", id.Int64())
}

func ServerId(id int32) zap.Field {
	return zap.Int32("serverid", id)
}

func GameId(name string) zap.Field {
	return zap.String("gameid", name)
}

func RegElapsedDay(day int) zap.Field {
	return zap.Int("rentday", day)
}
