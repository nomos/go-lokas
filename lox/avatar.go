package lox

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/util"
	"github.com/nomos/promise"
)

var _ lokas.IModel = (*Avatar)(nil)

func NewAvatar(id util.ID)*Avatar{
	ret:=&Avatar{
		Actor:NewActor(),
	}
	ret.SetId(id)
	return ret
}

var _ lokas.IActor = (*Avatar)(nil)

type Avatar struct {
	*Actor
	GameId string
	ServerId int32
}

func (this *Avatar) Deserialize(a lokas.IProcess) error {
	return nil
}

func (this *Avatar) Serialize(a lokas.IProcess) error {
	return nil
}

func (this *Avatar) Type() string {
	return "Avatar"
}

func (this *Avatar) Load(conf lokas.IConfig) error {
	panic("implement me")
}

func (this *Avatar) Unload() error {
	panic("implement me")
}

func (this *Avatar) Start() *promise.Promise {
	return promise.Resolve(nil)
}

func (this *Avatar) Stop() *promise.Promise {
	return promise.Resolve(nil)
}

func (this *Avatar) OnStart() error {
	return nil
}

func (this *Avatar) OnStop() error {
	return nil
}
