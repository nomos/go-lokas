package rox

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type Route struct {
	handler http.Handler

	buildOnly bool

	name string

	err error

	namedRoutes map[string]*Route

	routeConf
}

func (this *Route) SkipClean() bool {
	return this.skipClean
}

func (this *Route) Match(req *http.Request, match *RouteMatch) bool {
	if this.buildOnly || this.err != nil {
		return false
	}

	var matchErr error

	for _, m := range this.matchers {
		if matched := m.Match(req, match); !matched {
			if _, ok := m.(methodMatcher); ok {
				matchErr = ErrMethodMismatch
				continue
			}

			if match.MatchErr == ErrNotFound {
				match.MatchErr = nil
			}

			matchErr = nil
			return false
		}
	}

	if matchErr != nil {
		match.MatchErr = matchErr
		return false
	}

	if match.MatchErr == ErrMethodMismatch && this.handler != nil {

		match.MatchErr = nil

		match.Handler = this.handler
	}

	if match.Route == nil {
		match.Route = this
	}
	if match.Handler == nil {
		match.Handler = this.handler
	}
	if match.Vars == nil {
		match.Vars = make(map[string]string)
	}

	this.regexp.setMatch(req, match, this)
	return true
}

func (this *Route) GetError() error {
	return this.err
}

func (this *Route) BuildOnly() *Route {
	this.buildOnly = true
	return this
}

func (this *Route) Handler(handler http.Handler) *Route {
	if this.err == nil {
		this.handler = handler
	}
	return this
}

func (this *Route) HandlerFunc(f func(http.ResponseWriter, *http.Request)) *Route {
	return this.Handler(http.HandlerFunc(f))
}

func (this *Route) GetHandler() http.Handler {
	return this.handler
}

func (this *Route) Name(name string) *Route {
	if this.name != "" {
		this.err = fmt.Errorf("mux: route already has name %q, can't set %q",
			this.name, name)
	}
	if this.err == nil {
		this.name = name
		this.namedRoutes[name] = this
	}
	return this
}

func (this *Route) GetName() string {
	return this.name
}

type matcher interface {
	Match(*http.Request, *RouteMatch) bool
}

func (this *Route) addMatcher(m matcher) *Route {
	if this.err == nil {
		this.matchers = append(this.matchers, m)
	}
	return this
}

func (this *Route) addRegexpMatcher(tpl string, typ regexpType) error {
	if this.err != nil {
		return this.err
	}
	if typ == regexpTypePath || typ == regexpTypePrefix {
		if len(tpl) > 0 && tpl[0] != '/' {
			return fmt.Errorf("mux: path must start with a slash, got %q", tpl)
		}
		if this.regexp.path != nil {
			tpl = strings.TrimRight(this.regexp.path.template, "/") + tpl
		}
	}
	rr, err := newRouteRegexp(tpl, typ, routeRegexpOptions{
		strictSlash:    this.strictSlash,
		useEncodedPath: this.useEncodedPath,
	})
	if err != nil {
		return err
	}
	for _, q := range this.regexp.queries {
		if err = uniqueVars(rr.varsN, q.varsN); err != nil {
			return err
		}
	}
	if typ == regexpTypeHost {
		if this.regexp.path != nil {
			if err = uniqueVars(rr.varsN, this.regexp.path.varsN); err != nil {
				return err
			}
		}
		this.regexp.host = rr
	} else {
		if this.regexp.host != nil {
			if err = uniqueVars(rr.varsN, this.regexp.host.varsN); err != nil {
				return err
			}
		}
		if typ == regexpTypeQuery {
			this.regexp.queries = append(this.regexp.queries, rr)
		} else {
			this.regexp.path = rr
		}
	}
	this.addMatcher(rr)
	return nil
}

type headerMatcher map[string]string

func (m headerMatcher) Match(this *http.Request, match *RouteMatch) bool {
	return matchMapWithString(m, this.Header, true)
}

func (this *Route) Headers(pairs ...string) *Route {
	if this.err == nil {
		var headers map[string]string
		headers, this.err = mapFromPairsToString(pairs...)
		return this.addMatcher(headerMatcher(headers))
	}
	return this
}

type headerRegexMatcher map[string]*regexp.Regexp

func (m headerRegexMatcher) Match(this *http.Request, match *RouteMatch) bool {
	return matchMapWithRegex(m, this.Header, true)
}

func (this *Route) HeadersRegexp(pairs ...string) *Route {
	if this.err == nil {
		var headers map[string]*regexp.Regexp
		headers, this.err = mapFromPairsToRegex(pairs...)
		return this.addMatcher(headerRegexMatcher(headers))
	}
	return this
}

func (this *Route) Host(tpl string) *Route {
	this.err = this.addRegexpMatcher(tpl, regexpTypeHost)
	return this
}

type MatcherFunc func(*http.Request, *RouteMatch) bool

func (m MatcherFunc) Match(this *http.Request, match *RouteMatch) bool {
	return m(this, match)
}

func (this *Route) MatcherFunc(f MatcherFunc) *Route {
	return this.addMatcher(f)
}

type methodMatcher []string

func (m methodMatcher) Match(this *http.Request, match *RouteMatch) bool {
	return matchInArray(m, this.Method)
}

func (this *Route) Methods(methods ...string) *Route {
	for k, v := range methods {
		methods[k] = strings.ToUpper(v)
	}
	return this.addMatcher(methodMatcher(methods))
}

func (this *Route) Path(tpl string) *Route {
	this.err = this.addRegexpMatcher(tpl, regexpTypePath)
	return this
}

func (this *Route) PathPrefix(tpl string) *Route {
	this.err = this.addRegexpMatcher(tpl, regexpTypePrefix)
	return this
}

func (this *Route) Queries(pairs ...string) *Route {
	length := len(pairs)
	if length%2 != 0 {
		this.err = fmt.Errorf(
			"mux: number of parameters must be multiple of 2, got %v", pairs)
		return nil
	}
	for i := 0; i < length; i += 2 {
		if this.err = this.addRegexpMatcher(pairs[i]+"="+pairs[i+1], regexpTypeQuery); this.err != nil {
			return this
		}
	}

	return this
}

type schemeMatcher []string

func (m schemeMatcher) Match(this *http.Request, match *RouteMatch) bool {
	scheme := this.URL.Scheme

	if scheme == "" {
		if this.TLS == nil {
			scheme = "http"
		} else {
			scheme = "https"
		}
	}
	return matchInArray(m, scheme)
}

func (this *Route) Schemes(schemes ...string) *Route {
	for k, v := range schemes {
		schemes[k] = strings.ToLower(v)
	}
	if len(schemes) > 0 {
		this.buildScheme = schemes[0]
	}
	return this.addMatcher(schemeMatcher(schemes))
}

type BuildVarsFunc func(map[string]string) map[string]string

func (this *Route) BuildVarsFunc(f BuildVarsFunc) *Route {
	if this.buildVarsFunc != nil {

		old := this.buildVarsFunc
		this.buildVarsFunc = func(m map[string]string) map[string]string {
			return f(old(m))
		}
	} else {
		this.buildVarsFunc = f
	}
	return this
}

func (this *Route) Subrouter() *Router {

	router := &Router{routeConf: copyRouteConf(this.routeConf), namedRoutes: this.namedRoutes}
	this.addMatcher(router)
	return router
}

func (this *Route) URL(pairs ...string) (*url.URL, error) {
	if this.err != nil {
		return nil, this.err
	}
	values, err := this.prepareVars(pairs...)
	if err != nil {
		return nil, err
	}
	var scheme, host, path string
	queries := make([]string, 0, len(this.regexp.queries))
	if this.regexp.host != nil {
		if host, err = this.regexp.host.url(values); err != nil {
			return nil, err
		}
		scheme = "http"
		if this.buildScheme != "" {
			scheme = this.buildScheme
		}
	}
	if this.regexp.path != nil {
		if path, err = this.regexp.path.url(values); err != nil {
			return nil, err
		}
	}
	for _, q := range this.regexp.queries {
		var query string
		if query, err = q.url(values); err != nil {
			return nil, err
		}
		queries = append(queries, query)
	}
	return &url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     path,
		RawQuery: strings.Join(queries, "&"),
	}, nil
}

func (this *Route) URLHost(pairs ...string) (*url.URL, error) {
	if this.err != nil {
		return nil, this.err
	}
	if this.regexp.host == nil {
		return nil, errors.New("mux: route doesn't have a host")
	}
	values, err := this.prepareVars(pairs...)
	if err != nil {
		return nil, err
	}
	host, err := this.regexp.host.url(values)
	if err != nil {
		return nil, err
	}
	u := &url.URL{
		Scheme: "http",
		Host:   host,
	}
	if this.buildScheme != "" {
		u.Scheme = this.buildScheme
	}
	return u, nil
}

func (this *Route) URLPath(pairs ...string) (*url.URL, error) {
	if this.err != nil {
		return nil, this.err
	}
	if this.regexp.path == nil {
		return nil, errors.New("mux: route doesn't have a path")
	}
	values, err := this.prepareVars(pairs...)
	if err != nil {
		return nil, err
	}
	path, err := this.regexp.path.url(values)
	if err != nil {
		return nil, err
	}
	return &url.URL{
		Path: path,
	}, nil
}

func (this *Route) GetPathTemplate() (string, error) {
	if this.err != nil {
		return "", this.err
	}
	if this.regexp.path == nil {
		return "", errors.New("mux: route doesn't have a path")
	}
	return this.regexp.path.template, nil
}

func (this *Route) GetPathRegexp() (string, error) {
	if this.err != nil {
		return "", this.err
	}
	if this.regexp.path == nil {
		return "", errors.New("mux: route does not have a path")
	}
	return this.regexp.path.regexp.String(), nil
}

func (this *Route) GetQueriesRegexp() ([]string, error) {
	if this.err != nil {
		return nil, this.err
	}
	if this.regexp.queries == nil {
		return nil, errors.New("mux: route doesn't have queries")
	}
	queries := make([]string, 0, len(this.regexp.queries))
	for _, query := range this.regexp.queries {
		queries = append(queries, query.regexp.String())
	}
	return queries, nil
}

func (this *Route) GetQueriesTemplates() ([]string, error) {
	if this.err != nil {
		return nil, this.err
	}
	if this.regexp.queries == nil {
		return nil, errors.New("mux: route doesn't have queries")
	}
	queries := make([]string, 0, len(this.regexp.queries))
	for _, query := range this.regexp.queries {
		queries = append(queries, query.template)
	}
	return queries, nil
}

func (this *Route) GetMethods() ([]string, error) {
	if this.err != nil {
		return nil, this.err
	}
	for _, m := range this.matchers {
		if methods, ok := m.(methodMatcher); ok {
			return []string(methods), nil
		}
	}
	return nil, errors.New("mux: route doesn't have methods")
}

func (this *Route) GetHostTemplate() (string, error) {
	if this.err != nil {
		return "", this.err
	}
	if this.regexp.host == nil {
		return "", errors.New("mux: route doesn't have a host")
	}
	return this.regexp.host.template, nil
}

func (this *Route) prepareVars(pairs ...string) (map[string]string, error) {
	m, err := mapFromPairsToString(pairs...)
	if err != nil {
		return nil, err
	}
	return this.buildVars(m), nil
}

func (this *Route) buildVars(m map[string]string) map[string]string {
	if this.buildVarsFunc != nil {
		m = this.buildVarsFunc(m)
	}
	return m
}
