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

func copyRouteConf(r routeConf) routeConf {
	c := r

	if r.regexp.path != nil {
		c.regexp.path = copyRouteRegexp(r.regexp.path)
	}

	if r.regexp.host != nil {
		c.regexp.host = copyRouteRegexp(r.regexp.host)
	}

	c.regexp.queries = make([]*routeRegexp, 0, len(r.regexp.queries))
	for _, q := range r.regexp.queries {
		c.regexp.queries = append(c.regexp.queries, copyRouteRegexp(q))
	}

	c.matchers = make([]matcher, len(r.matchers))
	copy(c.matchers, r.matchers)

	return c
}

func copyRouteRegexp(r *routeRegexp) *routeRegexp {
	c := *r
	return &c
}

func (r *Router) Match(req *http.Request, match *RouteMatch) bool {
	for _, route := range r.routes {
		if route.Match(req, match) {

			if match.MatchErr == nil {
				for i := len(r.middlewares) - 1; i >= 0; i-- {
					mid := r.middlewares[i]
					if mid.matcher(req) {
						match.Handler = mid.middleware.Middleware(match.Handler)
					}
				}
			}
			return true
		}
	}

	if match.MatchErr == ErrMethodMismatch {
		if r.MethodNotAllowedHandler != nil {
			match.Handler = r.MethodNotAllowedHandler
			return true
		}

		return false
	}

	if r.NotFoundHandler != nil {
		match.Handler = r.NotFoundHandler
		match.MatchErr = ErrNotFound
		return true
	}

	match.MatchErr = ErrNotFound
	return false
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !r.skipClean {
		path := req.URL.Path
		if r.useEncodedPath {
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
	if r.Match(req, &match) {
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

func (r *Router) Get(name string) *Route {
	return r.namedRoutes[name]
}

func (r *Router) GetRoute(name string) *Route {
	return r.namedRoutes[name]
}

func (r *Router) StrictSlash(value bool) *Router {
	r.strictSlash = value
	return r
}

func (r *Router) SkipClean(value bool) *Router {
	r.skipClean = value
	return r
}

func (r *Router) UseEncodedPath() *Router {
	r.useEncodedPath = true
	return r
}

func (r *Router) NewRoute() *Route {

	route := &Route{routeConf: copyRouteConf(r.routeConf), namedRoutes: r.namedRoutes}
	r.routes = append(r.routes, route)
	return route
}

func (r *Router) Name(name string) *Route {
	return r.NewRoute().Name(name)
}

func (r *Router) Handle(path string, handler http.Handler) *Route {
	return r.NewRoute().Path(path).Handler(handler)
}

func (r *Router) HandleFunc(path string, f func(http.ResponseWriter,
	*http.Request)) *Route {
	return r.NewRoute().Path(path).HandlerFunc(f)
}

func (r *Router) Headers(pairs ...string) *Route {
	return r.NewRoute().Headers(pairs...)
}

func (r *Router) Host(tpl string) *Route {
	return r.NewRoute().Host(tpl)
}

func (r *Router) MatcherFunc(f MatcherFunc) *Route {
	return r.NewRoute().MatcherFunc(f)
}

func (r *Router) Methods(methods ...string) *Route {
	return r.NewRoute().Methods(methods...)
}

func (r *Router) Path(tpl string) *Route {
	return r.NewRoute().Path(tpl)
}

func (r *Router) PathPrefix(tpl string) *Route {
	return r.NewRoute().PathPrefix(tpl)
}

func (r *Router) Queries(pairs ...string) *Route {
	return r.NewRoute().Queries(pairs...)
}

func (r *Router) Schemes(schemes ...string) *Route {
	return r.NewRoute().Schemes(schemes...)
}

func (r *Router) BuildVarsFunc(f BuildVarsFunc) *Route {
	return r.NewRoute().BuildVarsFunc(f)
}

func (r *Router) Walk(walkFn WalkFunc) error {
	return r.walk(walkFn, []*Route{})
}

var SkipRouter = errors.New("skip this router")

type WalkFunc func(route *Route, router *Router, ancestors []*Route) error

func (r *Router) walk(walkFn WalkFunc, ancestors []*Route) error {
	for _, t := range r.routes {
		err := walkFn(t, r, ancestors)
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

func Vars(r *http.Request) map[string]string {
	if rv := r.Context().Value(varsKey); rv != nil {
		return rv.(map[string]string)
	}
	return nil
}

func CurrentRoute(r *http.Request) *Route {
	if rv := r.Context().Value(routeKey); rv != nil {
		return rv.(*Route)
	}
	return nil
}

func AddContext(r *http.Request, k interface{}, v interface{}) *http.Request {
	ctx := context.WithValue(r.Context(), k, v)
	return r.WithContext(ctx)
}

func requestWithVars(r *http.Request, vars map[string]string) *http.Request {
	ctx := context.WithValue(r.Context(), varsKey, vars)
	return r.WithContext(ctx)
}

func requestWithRoute(r *http.Request, route *Route) *http.Request {
	ctx := context.WithValue(r.Context(), routeKey, route)
	return r.WithContext(ctx)
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

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func methodNotAllowedHandler() http.Handler { return http.HandlerFunc(methodNotAllowed) }
