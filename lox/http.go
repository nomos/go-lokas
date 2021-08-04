package lox

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/rox"
	"github.com/nomos/go-lokas/network/httpserver"
	"github.com/nomos/go-lokas/promise"
	"net/http"
	"regexp"
	"sync"
)

var HttpCtor = httpCtor{}

type httpCtor struct {}

func (this httpCtor) Type()string{
	return "Http"
}

func (this httpCtor) Create() lokas.IModule {
	ret:=&Http{
		Actor:NewActor(),
		subRouters: map[string]*rox.Router{},
	}
	ret.SetType(this.Type())
	return ret
}

var _ lokas.IModule = (*Http)(nil)

type Http struct {
	*Actor
	Host string
	Port string
	subRouters map[string]*rox.Router
	httpServer *httpserver.HttpServer
	startPending *promise.Promise
	started bool
	mu sync.Mutex
}

func (this *Http) OnStart() error{
	return nil
}

func (this *Http) OnStop() error{
	return nil
}

func (this *Http) OnCreate() error {
	log.Info(this.Type()+" OnCreate")
	return nil
}

func (this *Http) OnDestroy() error {
	log.Info(this.Type()+" OnDestroy")
	return nil
}

func (this *Http) Load(conf lokas.IConfig)error {
	log.WithFields(log.Fields{
		"host":conf.Get("host"),
		"port":conf.Get("port"),
	}).Info("Http:LoadConfig")
	this.Host = conf.Get("host").(string)
	this.Port = conf.Get("port").(string)
	this.httpServer = httpserver.NewHttpServer()
	return nil
}

func (this *Http) Unload()error {
	return nil
}

func (this *Http) Start()*promise.Promise {
	log.Info(this.Type()+" Start")
	this.mu.Lock()
	defer this.mu.Unlock()
	if this.startPending==nil&&!this.started {
		this.startPending =  promise.Async(func(resolve func(interface{}), reject func(interface{})) {
			err:=this.httpServer.Start(this.Host+":"+this.Port)
			this.mu.Lock()
			defer this.mu.Unlock()
			if err != nil {
				this.startPending = nil
				reject(err)
				return
			}
			this.started = true
			this.startPending = nil
			resolve(nil)
		})
	} else if this.started {
		return promise.Resolve(nil)
	}
	return this.startPending
}

func (this *Http) Stop()*promise.Promise {
	this.mu.Lock()
	defer this.mu.Unlock()
	log.Info(this.Type()+" Stop")
	this.started = false
	this.httpServer.Stop()
	return promise.Resolve(nil)
}

func (this *Http) CreateHandlerFunc(f rox.Handler)func(http.ResponseWriter, *http.Request){
	return func(writer http.ResponseWriter, request *http.Request) {
		f(writer.(rox.ResponseWriter),request,this.GetProcess())
	}
}

func (this *Http) HandleFunc(path string, f rox.Handler) *rox.Route {
	if ro:=this.subRouters[path];ro!=nil {
		return ro.HandleFunc("",func(writer http.ResponseWriter, request *http.Request) {
			f(writer.(rox.ResponseWriter),request,this.GetProcess())
		})
	}
	return this.httpServer.Router.HandleFunc(path, func(writer http.ResponseWriter, request *http.Request) {
		f(writer.(rox.ResponseWriter),request,this.GetProcess())
	})
}

func (this *Http) MatchRouter(s string)*rox.Router{
	return this.httpServer.Router.MatcherFunc(func(request *http.Request, match *rox.RouteMatch) bool {
		return regexp.MustCompile(s).MatchString(request.URL.Path)
	}).Subrouter()
}

func (this *Http) Path(s string)*rox.Route{
	return this.httpServer.Router.Path(s)
}

func (this *Http) PathPrefix(s string)*rox.Route{
	return this.httpServer.Router.PathPrefix(s)
}

func (this *Http) PrefixOnly(prefix string,m ...rox.MiddleWare)*Http{
	mwf:=this.transformMiddleWares(m)
	this.httpServer.Router.When(func(req *http.Request) bool {
		return regexp.MustCompile("^"+prefix).MatchString(req.URL.Path)
	},mwf...)
	return this
}

func (this *Http) PrefixOnlyStrict(prefix string,m ...rox.MiddleWare)*Http{
	mwf:=this.transformMiddleWares(m)
	this.httpServer.Router.When(func(req *http.Request) bool {
		return regexp.MustCompile("^"+prefix+"/").MatchString(req.URL.Path)
	},mwf...)
	return this
}

func (this *Http) PrefixExcept(prefix string,m ...rox.MiddleWare)*Http{
	mwf:=this.transformMiddleWares(m)
	this.httpServer.Router.When(func(req *http.Request) bool {
		return !regexp.MustCompile("^"+prefix).MatchString(req.URL.Path)
	},mwf...)
	return this
}

func (this *Http) PrefixExceptStrict(prefix string,m ...rox.MiddleWare)*Http{
	mwf:=this.transformMiddleWares(m)
	this.httpServer.Router.When(func(req *http.Request) bool {
		return !regexp.MustCompile("^"+prefix+"/").MatchString(req.URL.Path)
	},mwf...)
	return this
}

func (this *Http) PathIn(p []string,m ...rox.MiddleWare)*Http{
	mwf:=this.transformMiddleWares(m)
	this.httpServer.Router.PathIn(p,mwf...)
	return this
}


func (this *Http) PathOnly(p string,m ...rox.MiddleWare)*Http{
	mwf:=this.transformMiddleWares(m)
	this.httpServer.Router.PathOnly(p,mwf...)
	return this
}

func (this *Http) PathExcept(p string,m ...rox.MiddleWare)*Http{
	mwf:=this.transformMiddleWares(m)
	this.httpServer.Router.PathExcept(p,mwf...)
	return this
}

func (this *Http) PathExcepts(p []string,m ...rox.MiddleWare)*Http{
	mwf:=this.transformMiddleWares(m)
	this.httpServer.Router.PathExcepts(p,mwf...)
	return this
}

func (this *Http) When(matcher rox.MiddlewareMatcher,m ...rox.MiddleWare)*Http {
	mwf:=this.transformMiddleWares(m)
	this.httpServer.Router.When(matcher,mwf...)
	return this
}

func (this *Http) Use(m ...rox.MiddleWare)*Http {
	mwf:=this.transformMiddleWares(m)
	this.httpServer.Router.Use(mwf...)
	return this
}

func (this *Http) transformMiddleWares(r []rox.MiddleWare)[]rox.MiddlewareFunc{
	ret:=make([]rox.MiddlewareFunc,0)
	for _,v:=range r {
		ret = append(ret, rox.InitMiddleWare(v,this.GetProcess()))
	}
	return ret
}