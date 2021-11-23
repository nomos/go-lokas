package lox

import (
	"github.com/nomos/go-lokas"
)


type CellManager struct {
	*Actor
	ActorContainer
	Blocks map[int64]Block
}

func (this *CellManager) Spawn() lokas.IActor {
	return nil
}

func (this *CellManager) Load(conf lokas.IConfig)error {
	return nil
}

func (this *CellManager) Unload()error {
	return nil
}

func (this *CellManager) OnStart()error {
	return nil
}

func (this *CellManager) OnStop()error {
	return nil
}