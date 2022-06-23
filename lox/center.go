package lox

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/rox"
)

var _ lokas.IModuleCtor = (*centerCtor)(nil)

type centerCtor struct{}

var CenterCtor = centerCtor{}

func (this centerCtor) Type() string {
	return "Center"
}

func (this centerCtor) Create() lokas.IModule {
	ret := &Center{
		Http: &Http{
			Actor:      NewActor(),
			subRouters: map[string]*rox.Router{},
		},
	}
	ret.SetType(this.Type())
	return ret
}

var _ lokas.IActor = &Center{}

type Center struct {
	*Http
	UseEtcd bool
}

func (this *Center) OnRegister() {

}

func (this *Center) OnStart() error {
	err := this.Http.OnStart()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Center) OnStop() error {
	err := this.Http.OnStop()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Center) OnCreate() error {
	err := this.Http.OnCreate()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Center) OnDestroy() error {
	err := this.Http.OnDestroy()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Center) Load(conf lokas.IConfig) error {
	this.UseEtcd = conf.Get("use_etcd").(bool)
	err := this.Http.Load(conf)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Center) Unload() error {
	err := this.Http.Unload()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Center) Start() error {
	return this.Http.Start()
}

func (this *Center) Stop() error {
	return this.Http.Stop()
}
