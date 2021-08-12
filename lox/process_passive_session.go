package lox

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/util"
)

type ProcessPassiveSession struct {
	*PassiveSession
}

func NewProcessPassiveSession(conn lokas.IConn, id util.ID, manager lokas.ISessionManager, opts ...PassiveSessionOption)*ProcessPassiveSession {
	ret:=&ProcessPassiveSession{NewPassiveSession(conn,id,manager,opts...)}
	return ret
}