package lox

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/util"
)

type ProcessActiveSession struct {
	*ActiveSession
}

func NewProcessActiveSession(conn lokas.IConn, id util.ID, manager lokas.ISessionManager, opts ...ActiveSessionOption)*ProcessActiveSession {
	ret:=&ProcessActiveSession{NewActiveSession(conn,id,manager,opts...)}
	return ret
}