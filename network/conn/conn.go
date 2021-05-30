package conn

import (
	"errors"
	"github.com/nomos/go-lokas"
	"net"
	"sync"
	"time"

	"github.com/nomos/go-lokas/network/internal/hub"
	"github.com/nomos/go-lokas/network/netstat"
)

type IOPumper interface {
	writePump(conn *Conn)
	readPump(conn *Conn)
}

type MessageChan interface {
	GetInChan() chan<- []byte
	GetOutChan() <-chan []byte
	Len() int
}

type Conn struct {
	net.Conn
	msgChan   MessageChan
	inChan    chan<- []byte
	context   *lokas.Context
	hub       *hub.Hub
	ioPumper  IOPumper
	wg        sync.WaitGroup
	Session   lokas.ISession
	UserData  interface{}
	ConnTime  time.Time
	stat      *netstat.NetStat
	done      chan struct{}
	closeOnce sync.Once
}

var errConnClosed = errors.New("connection already closed")

func newConn(c net.Conn, context *lokas.Context, msgChan MessageChan, hub *hub.Hub, done chan struct{}) *Conn {
	conn := &Conn{
		Conn:     c,
		msgChan:  msgChan,
		inChan:   msgChan.GetInChan(),
		context:  context,
		hub:      hub,
		ConnTime: time.Now(),
		done:     done,
	}
	if context.SessionCreator != nil {
		conn.Session = context.SessionCreator(conn)
	}
	if context.EnableStatistics {
		conn.stat = netstat.NewNetStat()
		conn.stat.Start()
	}
	return conn
}

func (this *Conn) SetUserData(userData interface{}) {
	this.UserData = userData
}

func (this *Conn) GetUserData() interface{} {
	return this.UserData
}

func (this *Conn) GetSession() lokas.ISession {
	return this.Session
}

func (this *Conn) GetConnTime() time.Time {
	return this.ConnTime
}

func (this *Conn) Activate() {
	if this.hub != nil {
		this.hub.ActivateConn(this)
	}
}

func (this *Conn) ServeIO() {
	this.wg.Add(2)
	go func() {
		this.ioPumper.writePump(this)
		this.wg.Done()
	}()

	go func() {
		if this.hub != nil {
			this.hub.AddConn(this)
		}
		this.Session.OnOpen(this)
		this.ioPumper.readPump(this)
		this.Session.OnClose(this)
		if this.hub != nil {
			this.hub.RemoveConn(this)
		}
		if this.stat != nil {
			this.stat.Stop()
		}
		this.wg.Done()
	}()
}

func (this *Conn) Write(data []byte) (int, error) {
	// if data == nil {
	// 	return 0, nil
	// }
	select {
	case this.inChan <- data:
		return len(data), nil
	case <-this.done:
		return 0, errConnClosed
	}
}

func (this *Conn) Close() error {
	var err error
	this.closeOnce.Do(func() {
		close(this.done)
		err = this.Conn.Close()
	})
	return err
}

// func (this *IConn) GraceClose() error {
// 	select {
// 	case this.inChan <- nil:
// 		return nil
// 	case <-this.done:
// 		return errConnClosed
// 	}
// 	return nil
// }

func (this *Conn) Wait() {
	this.wg.Wait()
}
