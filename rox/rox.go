package rox

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"regexp"
)

var (
	ErrMethodMismatch = errors.New("method is not allowed")

	ErrNotFound = errors.New("no matching route was found")
)

func NewRouter() *Router {
	return &Router{namedRoutes: make(map[string]*Route)}
}

type MiddlewareMatcher func(req *http.Request) bool

var defaultMiddlewareMatcher MiddlewareMatcher = func(req *http.Request) bool {
	return true
}

type middlewareWrapper struct {
	matcher    MiddlewareMatcher
	middleware middleware
}

type Router struct {
	NotFoundHandler http.Handler

	MethodNotAllowedHandler http.Handler

	routes []*Route

	namedRoutes map[string]*Route

	KeepContext bool

	middlewares []middlewareWrapper

	routeConf
}

type routeConf struct {
	useEncodedPath bool

	strictSlash bool

	skipClean bool

	regexp routeRegexpGroup

	matchers []matcher

	buildScheme string

	buildVarsFunc BuildVarsFunc
}

func copyRouteConf(this routeConf) routeConf {
	c := this

	if this.regexp.path != nil {
		c.regexp.path = copyRouteRegexp(this.regexp.path)
	}

	if this.regexp.host != nil {
		c.regexp.host = copyRouteRegexp(this.regexp.host)
	}

	c.regexp.queries = make([]*routeRegexp, 0, len(this.regexp.queries))
	for _, q := range this.regexp.queries {
		c.regexp.queries = append(c.regexp.queries, copyRouteRegexp(q))
	}

	c.matchers = make([]matcher, len(this.matchers))
	copy(c.matchers, this.matchers)

	return c
}

func copyRouteRegexp(this *routeRegexp) *routeRegexp {
	c := *this
	return &c
}

func (this *Router) Match(req *http.Request, match *RouteMatch) bool {
	for _, route := range this.routes {
		if route.Match(req, match) {

			if match.MatchErr == nil {
				for i := len(this.middlewares) - 1; i >= 0; i-- {
					mid := this.middlewares[i]
					if mid.matcher(req) {
						match.Handler = mid.middleware.Middleware(match.Handler)
					}
				}
			}
			return true
		}
	}

	if match.MatchErr == ErrMethodMismatch {
		if this.MethodNotAllowedHandler != nil {
			match.Handler = this.MethodNotAllowedHandler
			return true
		}

		return false
	}

	if this.NotFoundHandler != nil {
		match.Handler = this.NotFoundHandler
		match.MatchErr = ErrNotFound
		return true
	}

	match.MatchErr = ErrNotFound
	return false
}

func (this *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !this.skipClean {
		path := req.URL.Path
		if this.useEncodedPath {
			path = req.URL.EscapedPath()
		}

		if p := cleanPath(path); p != path {

			url := *req.URL
			url.Path = p
			p = url.String()

			w.Header().Set("Location", p)
			w.WriteHeader(http.StatusMovedPermanently)
			return
		}
	}
	var match RouteMatch
	var handler http.Handler
	if this.Match(req, &match) {
		handler = match.Handler
		req = requestWithVars(req, match.Vars)
		req = requestWithRoute(req, match.Route)
	}

	if handler == nil && match.MatchErr == ErrMethodMismatch {
		handler = methodNotAllowedHandler()
	}

	if handler == nil {
		handler = http.NotFoundHandler()
	}

	resp := NewResponseWriter(w)
	handler.ServeHTTP(resp, req)
	resp.WriteContent()
}

func (this *Router) Get(name string) *Route {
	return this.namedRoutes[name]
}

func (this *Router) GetRoute(name string) *Route {
	return this.namedRoutes[name]
}

func (this *Router) StrictSlash(value bool) *Router {
	this.strictSlash = value
	return this
}

func (this *Router) SkipClean(value bool) *Router {
	this.skipClean = value
	return this
}

func (this *Router) UseEncodedPath() *Router {
	this.useEncodedPath = true
	return this
}

func (this *Router) NewRoute() *Route {

	route := &Route{routeConf: copyRouteConf(this.routeConf), namedRoutes: this.namedRoutes}
	this.routes = append(this.routes, route)
	return route
}

func (this *Router) Name(name string) *Route {
	return this.NewRoute().Name(name)
}

func (this *Router) Handle(path string, handler http.Handler) *Route {
	return this.NewRoute().Path(path).Handler(handler)
}

func (this *Router) HandleFunc(path string, f func(http.ResponseWriter,
	*http.Request)) *Route {
	return this.NewRoute().Path(path).HandlerFunc(f)
}

func (this *Router) Headers(pairs ...string) *Route {
	return this.NewRoute().Headers(pairs...)
}

func (this *Router) Host(tpl string) *Route {
	return this.NewRoute().Host(tpl)
}

func (this *Router) MatcherFunc(f MatcherFunc) *Route {
	return this.NewRoute().MatcherFunc(f)
}

func (this *Router) Methods(methods ...string) *Route {
	return this.NewRoute().Methods(methods...)
}

func (this *Router) Path(tpl string) *Route {
	return this.NewRoute().Path(tpl)
}

func (this *Router) PathPrefix(tpl string) *Route {
	return this.NewRoute().PathPrefix(tpl)
}

func (this *Router) Queries(pairs ...string) *Route {
	return this.NewRoute().Queries(pairs...)
}

func (this *Router) Schemes(schemes ...string) *Route {
	return this.NewRoute().Schemes(schemes...)
}

func (this *Router) BuildVarsFunc(f BuildVarsFunc) *Route {
	return this.NewRoute().BuildVarsFunc(f)
}

func (this *Router) Walk(walkFn WalkFunc) error {
	return this.walk(walkFn, []*Route{})
}

var SkipRouter = errors.New("skip this router")

type WalkFunc func(route *Route, routethis *Router, ancestors []*Route) error

func (this *Router) walk(walkFn WalkFunc, ancestors []*Route) error {
	for _, t := range this.routes {
		err := walkFn(t, this, ancestors)
		if err == SkipRouter {
			continue
		}
		if err != nil {
			return err
		}
		for _, sr := range t.matchers {
			if h, ok := sr.(*Router); ok {
				ancestors = append(ancestors, t)
				err := h.walk(walkFn, ancestors)
				if err != nil {
					return err
				}
				ancestors = ancestors[:len(ancestors)-1]
			}
		}
		if h, ok := t.handler.(*Router); ok {
			ancestors = append(ancestors, t)
			err := h.walk(walkFn, ancestors)
			if err != nil {
				return err
			}
			ancestors = ancestors[:len(ancestors)-1]
		}
	}
	return nil
}

type RouteMatch struct {
	Route   *Route
	Handler http.Handler
	Vars    map[string]string

	MatchErr error
}

type contextKey int

const (
	varsKey contextKey = iota
	routeKey
)

func Vars(this *http.Request) map[string]string {
	if rv := this.Context().Value(varsKey); rv != nil {
		return rv.(map[string]string)
	}
	return nil
}

func CurrentRoute(this *http.Request) *Route {
	if rv := this.Context().Value(routeKey); rv != nil {
		return rv.(*Route)
	}
	return nil
}

func AddContext(this *http.Request, k interface{}, v interface{}) *http.Request {
	ctx := context.WithValue(this.Context(), k, v)
	return this.WithContext(ctx)
}

func requestWithVars(this *http.Request, vars map[string]string) *http.Request {
	ctx := context.WithValue(this.Context(), varsKey, vars)
	return this.WithContext(ctx)
}

func requestWithRoute(this *http.Request, route *Route) *http.Request {
	ctx := context.WithValue(this.Context(), routeKey, route)
	return this.WithContext(ctx)
}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)

	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}

	return np
}

func uniqueVars(s1, s2 []string) error {
	for _, v1 := range s1 {
		for _, v2 := range s2 {
			if v1 == v2 {
				return fmt.Errorf("mux: duplicated route variable %q", v2)
			}
		}
	}
	return nil
}

func checkPairs(pairs ...string) (int, error) {
	length := len(pairs)
	if length%2 != 0 {
		return length, fmt.Errorf(
			"mux: number of parameters must be multiple of 2, got %v", pairs)
	}
	return length, nil
}

func mapFromPairsToString(pairs ...string) (map[string]string, error) {
	length, err := checkPairs(pairs...)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, length/2)
	for i := 0; i < length; i += 2 {
		m[pairs[i]] = pairs[i+1]
	}
	return m, nil
}

func mapFromPairsToRegex(pairs ...string) (map[string]*regexp.Regexp, error) {
	length, err := checkPairs(pairs...)
	if err != nil {
		return nil, err
	}
	m := make(map[string]*regexp.Regexp, length/2)
	for i := 0; i < length; i += 2 {
		regex, err := regexp.Compile(pairs[i+1])
		if err != nil {
			return nil, err
		}
		m[pairs[i]] = regex
	}
	return m, nil
}

func matchInArray(arr []string, value string) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

func matchMapWithString(toCheck map[string]string, toMatch map[string][]string, canonicalKey bool) bool {
	for k, v := range toCheck {

		if canonicalKey {
			k = http.CanonicalHeaderKey(k)
		}
		if values := toMatch[k]; values == nil {
			return false
		} else if v != "" {

			valueExists := false
			for _, value := range values {
				if v == value {
					valueExists = true
					break
				}
			}
			if !valueExists {
				return false
			}
		}
	}
	return true
}

func matchMapWithRegex(toCheck map[string]*regexp.Regexp, toMatch map[string][]string, canonicalKey bool) bool {
	for k, v := range toCheck {

		if canonicalKey {
			k = http.CanonicalHeaderKey(k)
		}
		if values := toMatch[k]; values == nil {
			return false
		} else if v != nil {

			valueExists := false
			for _, value := range values {
				if v.MatchString(value) {
					valueExists = true
					break
				}
			}
			if !valueExists {
				return false
			}
		}
	}
	return true
}

func methodNotAllowed(w http.ResponseWriter, this *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func methodNotAllowedHandler() http.Handler { return http.HandlerFunc(methodNotAllowed) }
