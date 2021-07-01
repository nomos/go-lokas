package httpserver

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/rox"
	"net"
	"net/http"
	"sync"
)

type HttpServer struct {
	listener net.Listener
	closing  bool
	Router   *rox.Router
	wg       sync.WaitGroup
}

func NewHttpServer() *HttpServer {
	server := &HttpServer{Router: rox.NewRouter()}
	server.Router.StrictSlash(true)
	return server
}

func (this *HttpServer) Start(netAddr string) error {
	l, err := net.Listen("tcp", netAddr)
	if err != nil {
		return err
	}
	this.listener = l

	this.wg.Add(1)
	go func() {
		this.serve()
		this.wg.Done()
	}()
	return nil
}

func (this *HttpServer) Stop() {
	this.closing = true
	if this.listener != nil {
		this.listener.Close()
	}
	this.wg.Wait()
}

func (this *HttpServer) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	this.Router.HandleFunc(pattern, handler)
	this.Router.Use()
}

func (this *HttpServer) Handle(pattern string, handler http.Handler) {
	this.Router.Handle(pattern, handler)
}

func (this *HttpServer) serve() {
	err := http.Serve(this.listener, this.Router)
	if !this.closing && err != nil {
		log.Info("http serve error: " + err.Error())
	}
}
