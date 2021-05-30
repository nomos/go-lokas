package lox

import (
	"errors"
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas"
	"plugin"
)

type Creator func() lokas.IModule

var moduleRegistries map[string]Creator

func RegisterModule(creator Creator){
	name:=creator().Type()
	registerModule(name,creator)
}
func registerModule(s string,creator Creator){
	moduleRegistries[s] = creator
}

func NewModuleFromRegistry(s string)(lokas.IModule,error){
	if creator,ok:= moduleRegistries[s];ok{
		return creator(),nil
	}
	err:=errors.New("module not found:"+s)
	log.Error(err.Error())
	return nil,err
}

func RegisterModulesFromPlugin(path string)error{
	p, err := plugin.Open(path)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	m, err := p.Lookup("Mods")
	if err != nil {
		log.Error(err.Error())
		return err
	}
	modules := m.(func()map[string]Creator)()
	for k,v:=range modules {
		registerModule(k, v)
	}
	return nil
}