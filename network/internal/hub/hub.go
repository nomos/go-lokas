package hub

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/util"
	"sync"
	"sync/atomic"
	"time"
)

type broadcastMessage struct {
	sessionIds []util.ID
	data       []byte
}

type connChanData struct {
	conn lokas.IConn
	typ  int
}

const (
	chanTypeInit = iota + 1
	chanTypeActive
	chanTypeClose
)

type Hub struct {
	broadcastChan chan *broadcastMessage
	connChan      chan *connChanData
	activeConnNum int // the number of activeConns
	idleDuration  time.Duration
	initConns     map[lokas.IConn]bool    // connected but not authed IConn
	activeConns   map[util.ID]lokas.IConn // authed IConn
	done          chan struct{}
	closed        int32
	wg            sync.WaitGroup
}

func NewHub(idleDuration time.Duration) *Hub {
	hub := &Hub{
		connChan:      make(chan *connChanData, 100),
		broadcastChan: make(chan *broadcastMessage, 1000),
		idleDuration:  idleDuration,
		initConns:     make(map[lokas.IConn]bool),
		activeConns:   make(map[util.ID]lokas.IConn),
		done:          make(chan struct{}),
	}
	hub.wg.Add(1)
	go func() {
		hub.run()
		hub.wg.Done()
	}()
	return hub
}

func (this *Hub) AddConn(conn lokas.IConn) {
	closed := atomic.LoadInt32(&this.closed) == 1
	if !closed {
		if this.idleDuration > 0 {
			this.pushChanData(conn, chanTypeInit)
		} else {
			this.pushChanData(conn, chanTypeActive)
		}
	}
}

func (this *Hub) ActivateConn(conn lokas.IConn) {
	closed := atomic.LoadInt32(&this.closed) == 1
	if !closed && this.idleDuration > 0 {
		this.pushChanData(conn, chanTypeActive)
	}
}

func (this *Hub) RemoveConn(conn lokas.IConn) {
	closed := atomic.LoadInt32(&this.closed) == 1
	if !closed {
		this.pushChanData(conn, chanTypeClose)
	}
}

func (this *Hub) pushChanData(conn lokas.IConn, typ int) {
	select {
	case this.connChan <- &connChanData{conn, typ}:
	case <-this.done:
	}
}

func (this *Hub) GetActiveConnNum() int {
	return this.activeConnNum
}

func (this *Hub) Broadcast(sessionIds []util.ID, data []byte) {
	this.broadcastChan <- &broadcastMessage{sessionIds: sessionIds, data: data}
}

func (this *Hub) run() {
	var ticker = time.NewTicker(3 * time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case data := <-this.connChan:
			conn := data.conn
			switch data.typ {
			case chanTypeInit:
				this.initConns[conn] = true
			case chanTypeActive:
				if _, ok := this.initConns[conn]; ok {
					delete(this.initConns, conn)
				}
				this.activeConns[conn.GetSession().GetId()] = conn
			case chanTypeClose:
				delete(this.initConns, conn)
				delete(this.activeConns, conn.GetSession().GetId())
			}
		case message := <-this.broadcastChan:
			if len(message.sessionIds) == 0 {
				for _, conn := range this.activeConns {
					conn.Write(message.data)
				}
			} else {
				for _, id := range message.sessionIds {
					conn := this.activeConns[id]
					if conn != nil {
						conn.Write(message.data)
					}
				}
			}
		case <-ticker.C: // check active state
			this.activeConnNum = len(this.activeConns)
			if this.idleDuration > 0 && len(this.initConns) > 0 {
				now := time.Now()
				for conn := range this.initConns {
					if now.Sub(conn.GetConnTime()) > this.idleDuration {
						delete(this.initConns, conn)
						conn.Close()
					}
				}
			}
		case <-this.done:
			this.clear()
			return
		}
	}
}

func (this *Hub) Stop() {
	if atomic.CompareAndSwapInt32(&this.closed, 0, 1) {
		close(this.done)
		this.wg.Wait()
	}
}

func (this *Hub) clear() {
	conns := make(map[lokas.IConn]bool)
	n := len(this.connChan)
	for i := 0; i < n; i++ {
		data := <-this.connChan
		if data.typ != chanTypeClose {
			conns[data.conn] = true
		}
	}
	for conn := range this.initConns {
		conns[conn] = true
	}
	for _, conn := range this.activeConns {
		conns[conn] = true
	}
	for conn, _ := range conns {
		conn.Close()
	}
	for conn := range conns {
		conn.Wait()
	}
}
