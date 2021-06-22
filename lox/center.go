package lox

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/promise"
)

var _ lokas.IModuleCtor = (*centerCtor)(nil)

type centerCtor struct {}

var CenterCtor = centerCtor{}

func (this centerCtor) Type()string{
	return "Center"
}

func (this centerCtor) Create()lokas.IModule {
	ret:= &Center{
		Actor:NewActor(),
	}
	ret.SetType(this.Type())
	return ret
}

var _ lokas.IActor = &Center{}

type Center struct {
	*Actor
	UseEtcd        bool
}

func (this *Center) OnRegister() {
	panic("implement me")
}

func (this *Center) OnStart() error{
	panic("implement me")
}

func (this *Center) OnStop() error{
	panic("implement me")
}

func (this *Center) OnCreate() error {
	panic("implement me")
}

func (this *Center) OnDestroy() error {
	panic("implement me")
}

func (this *Center) Load(conf lokas.IConfig)error {
	this.UseEtcd = conf.Get("use_etcd").(bool)
	return nil
}

func (this *Center) Unload()error {
	return nil
}

func (this *Center) Start()*promise.Promise {
	return nil
}

func (this *Center) Stop()*promise.Promise {
	return nil
}