package lox

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
)

type GameHandler struct {
	Serializer   func(avatar lokas.IActor, process lokas.IProcess) error
	Initializer  func(avatar lokas.IActor, process lokas.IProcess) error
	Deserializer func(avatar lokas.IActor, process lokas.IProcess) error
	Updater      func(avatar lokas.IActor, process lokas.IProcess) error
	MsgDelegator func(avatar lokas.IActor, actorId util.ID, transId uint32, msg protocol.ISerializable) (protocol.ISerializable, error)
}

func (this *GameHandler) GetSerializer() func(avatar lokas.IActor, process lokas.IProcess) error {
	return this.Serializer
}

func (this *GameHandler) GetInitializer() func(avatar lokas.IActor, process lokas.IProcess) error {
	return this.Initializer
}

func (this *GameHandler) GetDeserializer() func(avatar lokas.IActor, process lokas.IProcess) error {
	return this.Deserializer
}

func (this *GameHandler) GetUpdater() func(avatar lokas.IActor, process lokas.IProcess) error {
	return this.Updater
}

func (this *GameHandler) GetMsgDelegator() func(avatar lokas.IActor, actorId util.ID, transId uint32, msg protocol.ISerializable) (protocol.ISerializable, error) {
	return this.MsgDelegator
}
