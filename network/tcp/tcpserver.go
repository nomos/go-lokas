package tcp

import (
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/network/conn"
	"github.com/nomos/go-lokas/network/internal/hub"
	"github.com/nomos/go-lokas/util"
	"net"
	"runtime"
	"strings"
)

type Server struct {
	Context  *lokas.Context
	hub      *hub.Hub
	listener net.Listener
}

// NewServer create a new tcp server
func NewServer(context *lokas.Context) *Server {
	if context == nil || context.SessionCreator == nil || context.Splitter == nil {
		panic("tcpserver.NewServer: context.SessionCreator is nil or context.Splitter is nil")
	}

	server := &Server{
		Context: context,
		hub:     hub.NewHub(context.IdleTimeAfterOpen),
	}
	return server
}

// Start start tcp server
func (this *Server) Start(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	this.listener = l
	go this.serve()
	return nil
}

func (this *Server) serve() {
	l := this.listener
	for {
		c, err := l.Accept()
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				log.Infof("tcpserver: temporary Accept() error, %s", err.Error())
				runtime.Gosched()
				continue
			}
			// theres no direct way to detect this error because it is not exposed
			if !strings.Contains(err.Error(), "use of closed network connection") {
				log.Infof("tcpserver: listener.Accept() error, %s", err.Error())
			}
			break
		}
		conn := conn.NewTcpConn(c, this.Context, this.hub)
		conn.ServeIO()
	}
}

// Stop stop tcp server
func (this *Server) Stop() {
	if this.listener != nil {
		this.listener.Close()
	}
	this.hub.Stop()
}

// Broadcast broadcast data to all active connections
func (this *Server) Broadcast(sessionIds []util.ID, data []byte) {
	this.hub.Broadcast(sessionIds, data)
}

// GetActiveConnNum get count of active connections
func (this *Server) GetActiveConnNum() int {
	return this.hub.GetActiveConnNum()
}