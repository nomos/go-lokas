package cmds

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/util/promise"
)

type IConsole interface {
	log.Hook
	log.ILogger
	Write(p []byte)(int,error)
	Clear()
}

type ICommandSender interface {
	SendCmd(string)
	OnSelect()
	OnDeselect()
}

type ICommand interface {
	Name()string
	SetConsole(IConsole)
	ConsoleExec(param *ParamsValue,console IConsole)*promise.Promise
	ExecWithConsole(console IConsole,params... string)*promise.Promise
	Exec(params... string)*promise.Promise
	Tips()string
}