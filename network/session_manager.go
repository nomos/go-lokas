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
	defer this.sessionsMu.Unlock()
	this.sessions[id] = session
}

func (this *DefaultSessionManager) ResetSession(id util.ID, session lokas.ISession) {
	if !this.safeMode {
		delete(this.sessions,session.GetId())
		this.sessions[id] = session
		return
	}
	this.sessionsMu.Lock()
	defer this.sessionsMu.Unlock()
	delete(this.sessions,session.GetId())
	this.sessions[id] = session
}

func (this *DefaultSessionManager) RemoveSession(id util.ID) {
	if !this.safeMode {
		delete(this.sessions, id)
		return
	}
	this.sessionsMu.Lock()
	defer this.sessionsMu.Unlock()
	delete(this.sessions, id)
}

func (this *DefaultSessionManager) GetSession(id util.ID) lokas.ISession {
	if !this.safeMode {
		return this.sessions[id]
	}
	this.sessionsMu.RLock()
	defer this.sessionsMu.RUnlock()
	session := this.sessions[id]
	return session
}

func (this *DefaultSessionManager) GetRoundSession() (lokas.ISession, bool) {
	var keys = make([]util.ID, 0)
	this.sessionsMu.RLock()
	defer this.sessionsMu.RUnlock()
	for key, _ := range this.sessions {
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		return nil, false
	}
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	index := r.Intn(len(keys))
	session := this.sessions[keys[index]]

	return session, true
}

func (this *DefaultSessionManager) Range(f func(id util.ID, session lokas.ISession) bool) {
	if this.safeMode {
		this.sessionsMu.RLock()
		defer this.sessionsMu.RUnlock()
	}
	for id, session := range this.sessions {
		if f(id, session) {
			break
		}
	}
}

func (this *DefaultSessionManager) GetSessionCount() int {
	if this.safeMode {
		this.sessionsMu.RLock()
		defer this.sessionsMu.RUnlock()
	}
	count := len(this.sessions)
	return count
}

func (this *DefaultSessionManager) Clear() {
	if this.safeMode {
		this.sessionsMu.RLock()
		defer this.sessionsMu.RUnlock()
	}
	for id,sess := range this.sessions {
		sess.GetConn().Close()
		delete(this.sessions, id)
	}
}


