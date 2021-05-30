package network

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/util"
	"math/rand"
	"sync"
	"time"
)


type DefaultSessionManager struct {
	sessions   map[util.ID]lokas.ISession // key may not be equal to session.Id(), let application layer decide
	sessionsMu sync.RWMutex
	safeMode   bool
}

func NewDefaultSessionManager(safeMode bool) *DefaultSessionManager {
	return &DefaultSessionManager{
		sessions: make(map[util.ID]lokas.ISession),
		safeMode: safeMode,
	}
}

func (this *DefaultSessionManager) AddSession(id util.ID, session lokas.ISession) {
	if !this.safeMode {
		this.sessions[id] = session
		return
	}
	this.sessionsMu.Lock()
	this.sessions[id] = session
	this.sessionsMu.Unlock()
}

func (this *DefaultSessionManager) RemoveSession(id util.ID) {
	if !this.safeMode {
		delete(this.sessions, id)
		return
	}
	this.sessionsMu.Lock()
	delete(this.sessions, id)
	this.sessionsMu.Unlock()
}

func (this *DefaultSessionManager) GetSession(id util.ID) lokas.ISession {
	if !this.safeMode {
		return this.sessions[id]
	}
	this.sessionsMu.RLock()
	session := this.sessions[id]
	this.sessionsMu.RUnlock()
	return session
}

func (this *DefaultSessionManager) GetRoundSession() (lokas.ISession, bool) {
	var keys = make([]util.ID, 0)
	this.sessionsMu.RLock()
	for key, _ := range this.sessions {
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		this.sessionsMu.RUnlock()
		return nil, false
	}
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	index := r.Intn(len(keys))
	session := this.sessions[keys[index]]
	this.sessionsMu.RUnlock()
	return session, true
}

func (this *DefaultSessionManager) Range(f func(id util.ID, session lokas.ISession) bool) {
	if this.safeMode {
		this.sessionsMu.RLock()
	}
	for id, session := range this.sessions {
		if f(id, session) {
			break
		}
	}
	if this.safeMode {
		this.sessionsMu.RUnlock()
	}
}

func (this *DefaultSessionManager) GetSessionCount() int {
	if this.safeMode {
		this.sessionsMu.RLock()
	}
	count := len(this.sessions)
	if this.safeMode {
		this.sessionsMu.RUnlock()
	}
	return count
}

func (this *DefaultSessionManager) Clear() {
	if this.safeMode {
		this.sessionsMu.RLock()
	}
	for id,sess := range this.sessions {
		sess.GetConn().Close()
		delete(this.sessions, id)
	}
	if this.safeMode {
		this.sessionsMu.RUnlock()
	}
}


