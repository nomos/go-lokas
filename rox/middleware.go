package rox

import (
	"github.com/nomos/go-lokas"
	"net/http"
	"strings"
)

type MiddleWare func (w ResponseWriter, r *http.Request, a lokas.IProcess,next http.Handler)

func InitMiddleWare(f func (w ResponseWriter, r *http.Request, a lokas.IProcess,next http.Handler),a lokas.IProcess)MiddlewareFunc{
	return func (next http.Handler)http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			f(writer.(ResponseWriter),request,a,next)
		})
	}
}

func CreateMiddleWare(f func (w ResponseWriter, r *http.Request, a lokas.IProcess,next http.Handler))func (w ResponseWriter, r *http.Request, a lokas.IProcess,next http.Handler) {
	return f
}

// MiddlewareFunc is a function which receives an http.Handler and returns another http.Handlethis.
// Typically, the returned handler is a closure which does something with the http.ResponseWriter and http.Request passed
// to it, and then calls the handler passed as parameter to the MiddlewareFunc.
type MiddlewareFunc func(http.Handler) http.Handler

// middleware interface is anything which implements a MiddlewareFunc named Middleware.
type middleware interface {
	Middleware(handler http.Handler) http.Handler
}

// Middleware allows MiddlewareFunc to implement the middleware interface.
func (mw MiddlewareFunc) Middleware(handler http.Handler) http.Handler {
	return mw(handler)
}

// Use appends a MiddlewareFunc to the chain. Middleware can be used to intercept or otherwise modify requests and/or responses, and are executed in the order that they are applied to the Routethis.
func (this *Router) Use(mwf ...MiddlewareFunc) {
	for _, fn := range mwf {
		this.middlewares = append(this.middlewares, middlewareWrapper{
			matcher:    defaultMiddlewareMatcher,
			middleware: fn,
		})
	}
}

func (this *Router) When(match MiddlewareMatcher,mwf ...MiddlewareFunc) {
	for _, fn := range mwf {
		this.middlewares = append(this.middlewares, middlewareWrapper{
			matcher:    match,
			middleware: fn,
		})
	}
}

func (this *Router) PathIn(p []string,mwf ...MiddlewareFunc) {
	for _, fn := range mwf {
		this.middlewares = append(this.middlewares, middlewareWrapper{
			matcher: func(req *http.Request) bool {
				for _,v:=range p {
					if v==req.URL.Path {
						return true
					}
				}
				return false
			},
			middleware: fn,
		})
	}
}

func (this *Router) PathOnly(p string,mwf ...MiddlewareFunc) {
	for _, fn := range mwf {
		this.middlewares = append(this.middlewares, middlewareWrapper{
			matcher: func(req *http.Request) bool {
				return req.URL.Path == p
			},
			middleware: fn,
		})
	}
}

func (this *Router) PathExcept(p string,mwf ...MiddlewareFunc) {
	for _, fn := range mwf {
		this.middlewares = append(this.middlewares, middlewareWrapper{
			matcher: func(req *http.Request) bool {
				return req.URL.Path != p
			},
			middleware: fn,
		})
	}
}

func (this *Router) PathExcepts(p []string,mwf ...MiddlewareFunc) {
	for _, fn := range mwf {
		this.middlewares = append(this.middlewares, middlewareWrapper{
			matcher: func(req *http.Request) bool {
				for _,v:=range p {
					if v==req.URL.Path {
						return false
					}
				}
				return true
			},
			middleware: fn,
		})
	}
}




// useInterface appends a middleware to the chain. Middleware can be used to intercept or otherwise modify requests and/or responses, and are executed in the order that they are applied to the Routethis.
func (this *Router) useInterface(mw middleware) {
	this.middlewares = append(this.middlewares, middlewareWrapper{
		matcher:    defaultMiddlewareMatcher,
		middleware: mw,
	})
}

// CORSMethodMiddleware automatically sets the Access-Control-Allow-Methods response header
// on requests for routes that have an OPTIONS method matcher to all the method matchers on
// the route. Routes that do not explicitly handle OPTIONS requests will not be processed
// by the middleware. See examples for usage.
func CORSMethodMiddleware(this *Router) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			allMethods, err := getAllMethodsForRoute(this, req)
			if err == nil {
				for _, v := range allMethods {
					if v == http.MethodOptions {
						w.Header().Set("Access-Control-Allow-Methods", strings.Join(allMethods, ","))
					}
				}
			}

			next.ServeHTTP(w, req)
		})
	}
}

// getAllMethodsForRoute returns all the methods from method matchers matching a given
// request.
func getAllMethodsForRoute(this *Router, req *http.Request) ([]string, error) {
	var allMethods []string

	for _, route := range this.routes {
		var match RouteMatch
		if route.Match(req, &match) || match.MatchErr == ErrMethodMismatch {
			methods, err := route.GetMethods()
			if err != nil {
				return nil, err
			}

			allMethods = append(allMethods, methods...)
		}
	}

	return allMethods, nil
}
