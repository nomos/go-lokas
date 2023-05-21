package rpc

import (
	"github.com/nomos/go-lokas/cmds"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/lox"
)

type rpcFunc func(cmd *lox.AdminCommand, params *cmds.ParamsValue, logger log.ILogger) ([]byte, error)

var rpcHandlers = map[string]rpcFunc{}

func RegisterAdminFunc(name string, f rpcFunc) {
	log.Infof("RegisterAdminFunc", name)
	rpcHandlers[name] = f
}

func GetRpcHandler(s string) rpcFunc {
	return rpcHandlers[s]
}

func GetRpcHandlers() map[string]rpcFunc {
	return rpcHandlers
}
