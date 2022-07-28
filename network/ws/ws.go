package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/network/conn"
	"github.com/nomos/go-lokas/network/httpserver"
	"github.com/nomos/go-lokas/network/internal/hub"
	"github.com/nomos/go-lokas/util"
	"go.uber.org/zap"
)

type WsServer struct {
	lokas.Server
	Context    *lokas.Context
	hub        *hub.Hub
	upgrader   *websocket.Upgrader
	httpServer *httpserver.HttpServer
}

// NewServer create a new websocket server
func NewWsServer(context *lokas.Context) *WsServer {
	if context == nil || context.SessionCreator == nil {
		panic("wsserver.NewServer: context is nil or context.SessionCreator is nil")
	}
	server := &WsServer{
		Context: context,
		hub:     hub.NewHub(context.IdleTimeAfterOpen),
	}
	server.upgrader = &websocket.Upgrader{
		ReadBufferSize:  1024 * 2,
		WriteBufferSize: 1024 * 4,
		CheckOrigin:     func(r *http.Request) bool { return true }, // disable check
	}
	if context.ReadBufferSize > 0 {
		server.upgrader.ReadBufferSize = context.ReadBufferSize
	}
	if context.WriteBufferSize > 0 {
		server.upgrader.WriteBufferSize = context.WriteBufferSize
	}
	return server
}

// Start start websocket server, and start default http server if addr is not empty
func (this *WsServer) Start(addr string) error {
	if len(addr) > 0 {
		httpServer := httpserver.NewHttpServer()
		err := httpServer.Start(addr)
		if err != nil {
			return err
		}
		httpServer.Handle("/ws", this)
		this.httpServer = httpServer
	}
	return nil
}

// ServeHTTP serve http request
func (this *WsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := this.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Info("wsserver.ServeHTTP upgrade error: %s", zap.String("err", err.Error()))
		return
	}
	conn := conn.NewWsConn(c, this.Context, this.hub)
	conn.ServeIO()
}

// Stop stop websocket server, and the underline default http server
func (this *WsServer) Stop() {
	if this.httpServer != nil {
		this.httpServer.Stop()
	}
	this.hub.Stop()
}

// Broadcast broadcast data to all active connections
func (this *WsServer) Broadcast(sessionIds []util.ID, data []byte) {
	this.hub.Broadcast(sessionIds, data)
}

// GetActiveConnNum get count of active connections
func (this *WsServer) GetActiveConnNum() int {
	return this.hub.GetActiveConnNum()
}
