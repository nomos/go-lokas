package lox

import (
	"encoding/json"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
)

type ProcessActiveSession struct {
	*ActiveSession
	verified bool
}

func processActiveSessionCreator(id util.ProcessId, p *Proxy) func(conn lokas.IConn) lokas.ISession {
	return func(conn lokas.IConn) lokas.ISession {
		sess := p.ActiveSessions.GetSession(id.Snowflake()).(*ProcessActiveSession)
		if sess == nil {
			sess = NewProcessActiveSession(conn, id.Snowflake(), p.ActiveSessions)
		}
		sess.OnOpenFunc = sess.onOpen
		sess.MsgHandler = sess.msgHandler
		sess.Conn = conn
		return sess
	}
}

func NewProcessActiveSession(conn lokas.IConn, id util.ID, manager lokas.ISessionManager, opts ...ActiveSessionOption) *ProcessActiveSession {
	s:=&ProcessActiveSession{
		ActiveSession:NewActiveSession(conn,id,manager,opts...),
	}
	return s
}

func (this *ProcessActiveSession) onOpen(conn lokas.IConn){
	d,_:=json.Marshal(&processHandShake{Id: this.GetId()})
	hs:=&protocol.HandShake{
		Data: d,
	}
	data,_:=protocol.MarshalMessage(0,hs,protocol.BINARY)
	this.Conn.Write(data)
}

func (this *ProcessActiveSession) msgHandler(msg *protocol.BinaryMessage){
	id,_:=msg.GetId()
	switch id {
	case protocol.TAG_HandShake:
		if this.OnVerified!=nil {
			this.verified = true
			this.OnVerified(this.Conn)
		}
	default:

	}
	log.Warnf()
}

func (this *ProcessActiveSession) onVerified(){
}

