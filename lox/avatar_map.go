package lox

import (
	"github.com/nomos/go-lokas/util"
)

type AccState uint32

const (
	ACC_NORMAL AccState = 0
	ACC_MUTE AccState = 1
	ACC_FROST AccState = 8
	ACC_FORBIDDEN AccState = 9
)

type AvatarMap struct {
	Id util.ID `bson:"_id"`
	UserId util.ID
	GameId string
	ServerId int32
	UserName string
}

