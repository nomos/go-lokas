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

func (r *Route) SkipClean() bool {
	return r.skipClean
}

func (r *Route) Match(req *http.Request, match *RouteMatch) bool {
	if r.buildOnly || r.err != nil {
		return false
	}

	var matchErr error

	for _, m := range r.matchers {
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

	if match.MatchErr == ErrMethodMismatch && r.handler != nil {

		match.MatchErr = nil

		match.Handler = r.handler
	}

	if match.Route == nil {
		match.Route = r
	}
	if match.Handler == nil {
		match.Handler = r.handler
	}
	if match.Vars == nil {
		match.Vars = make(map[string]string)
	}

	r.regexp.setMatch(req, match, r)
	return true
}

func (r *Route) GetError() error {
	return r.err
}

func (r *Route) BuildOnly() *Route {
	r.buildOnly = true
	return r
}

func (r *Route) Handler(handler http.Handler) *Route {
	if r.err == nil {
		r.handler = handler
	}
	return r
}

func (r *Route) HandlerFunc(f func(http.ResponseWriter, *http.Request)) *Route {
	return r.Handler(http.HandlerFunc(f))
}

func (r *Route) GetHandler() http.Handler {
	return r.handler
}

func (r *Route) Name(name string) *Route {
	if r.name != "" {
		r.err = fmt.Errorf("mux: route already has name %q, can't set %q",
			r.name, name)
	}
	if r.err == nil {
		r.name = name
		r.namedRoutes[name] = r
	}
	return r
}

func (r *Route) GetName() string {
	return r.name
}

type matcher interface {
	Match(*http.Request, *RouteMatch) bool
}

func (r *Route) addMatcher(m matcher) *Route {
	if r.err == nil {
		r.matchers = append(r.matchers, m)
	}
	return r
}

func (r *Route) addRegexpMatcher(tpl string, typ regexpType) error {
	if r.err != nil {
		return r.err
	}
	if typ == regexpTypePath || typ == regexpTypePrefix {
		if len(tpl) > 0 && tpl[0] != '/' {
			return fmt.Errorf("mux: path must start with a slash, got %q", tpl)
		}
		if r.regexp.path != nil {
			tpl = strings.TrimRight(r.regexp.path.template, "/") + tpl
		}
	}
	rr, err := newRouteRegexp(tpl, typ, routeRegexpOptions{
		strictSlash:    r.strictSlash,
		useEncodedPath: r.useEncodedPath,
	})
	if err != nil {
		return err
	}
	for _, q := range r.regexp.queries {
		if err = uniqueVars(rr.varsN, q.varsN); err != nil {
			return err
		}
	}
	if typ == regexpTypeHost {
		if r.regexp.path != nil {
			if err = uniqueVars(rr.varsN, r.regexp.path.varsN); err != nil {
				return err
			}
		}
		r.regexp.host = rr
	} else {
		if r.regexp.host != nil {
			if err = uniqueVars(rr.varsN, r.regexp.host.varsN); err != nil {
				return err
			}
		}
		if typ == regexpTypeQuery {
			r.regexp.queries = append(r.regexp.queries, rr)
		} else {
			r.regexp.path = rr
		}
	}
	r.addMatcher(rr)
	return nil
}

type headerMatcher map[string]string

func (m headerMatcher) Match(r *http.Request, match *RouteMatch) bool {
	return matchMapWithString(m, r.Header, true)
}

func (r *Route) Headers(pairs ...string) *Route {
	if r.err == nil {
		var headers map[string]string
		headers, r.err = mapFromPairsToString(pairs...)
		return r.addMatcher(headerMatcher(headers))
	}
	return r
}

type headerRegexMatcher map[string]*regexp.Regexp

func (m headerRegexMatcher) Match(r *http.Request, match *RouteMatch) bool {
	return matchMapWithRegex(m, r.Header, true)
}

func (r *Route) HeadersRegexp(pairs ...string) *Route {
	if r.err == nil {
		var headers map[string]*regexp.Regexp
		headers, r.err = mapFromPairsToRegex(pairs...)
		return r.addMatcher(headerRegexMatcher(headers))
	}
	return r
}

func (r *Route) Host(tpl string) *Route {
	r.err = r.addRegexpMatcher(tpl, regexpTypeHost)
	return r
}

type MatcherFunc func(*http.Request, *RouteMatch) bool

func (m MatcherFunc) Match(r *http.Request, match *RouteMatch) bool {
	return m(r, match)
}

func (r *Route) MatcherFunc(f MatcherFunc) *Route {
	return r.addMatcher(f)
}

type methodMatcher []string

func (m methodMatcher) Match(r *http.Request, match *RouteMatch) bool {
	return matchInArray(m, r.Method)
}

func (r *Route) Methods(methods ...string) *Route {
	for k, v := range methods {
		methods[k] = strings.ToUpper(v)
	}
	return r.addMatcher(methodMatcher(methods))
}

func (r *Route) Path(tpl string) *Route {
	r.err = r.addRegexpMatcher(tpl, regexpTypePath)
	return r
}

func (r *Route) PathPrefix(tpl string) *Route {
	r.err = r.addRegexpMatcher(tpl, regexpTypePrefix)
	return r
}

func (r *Route) Queries(pairs ...string) *Route {
	length := len(pairs)
	if length%2 != 0 {
		r.err = fmt.Errorf(
			"mux: number of parameters must be multiple of 2, got %v", pairs)
		return nil
	}
	for i := 0; i < length; i += 2 {
		if r.err = r.addRegexpMatcher(pairs[i]+"="+pairs[i+1], regexpTypeQuery); r.err != nil {
			return r
		}
	}

	return r
}

type schemeMatcher []string

func (m schemeMatcher) Match(r *http.Request, match *RouteMatch) bool {
	scheme := r.URL.Scheme

	if scheme == "" {
		if r.TLS == nil {
			scheme = "http"
		} else {
			scheme = "https"
		}
	}
	return matchInArray(m, scheme)
}

func (r *Route) Schemes(schemes ...string) *Route {
	for k, v := range schemes {
		schemes[k] = strings.ToLower(v)
	}
	if len(schemes) > 0 {
		r.buildScheme = schemes[0]
	}
	return r.addMatcher(schemeMatcher(schemes))
}

type BuildVarsFunc func(map[string]string) map[string]string

func (r *Route) BuildVarsFunc(f BuildVarsFunc) *Route {
	if r.buildVarsFunc != nil {

		old := r.buildVarsFunc
		r.buildVarsFunc = func(m map[string]string) map[string]string {
			return f(old(m))
		}
	} else {
		r.buildVarsFunc = f
	}
	return r
}

func (r *Route) Subrouter() *Router {

	router := &Router{routeConf: copyRouteConf(r.routeConf), namedRoutes: r.namedRoutes}
	r.addMatcher(router)
	return router
}

func (r *Route) URL(pairs ...string) (*url.URL, error) {
	if r.err != nil {
		return nil, r.err
	}
	values, err := r.prepareVars(pairs...)
	if err != nil {
		return nil, err
	}
	var scheme, host, path string
	queries := make([]string, 0, len(r.regexp.queries))
	if r.regexp.host != nil {
		if host, err = r.regexp.host.url(values); err != nil {
			return nil, err
		}
		scheme = "http"
		if r.buildScheme != "" {
			scheme = r.buildScheme
		}
	}
	if r.regexp.path != nil {
		if path, err = r.regexp.path.url(values); err != nil {
			return nil, err
		}
	}
	for _, q := range r.regexp.queries {
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

func (r *Route) URLHost(pairs ...string) (*url.URL, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.regexp.host == nil {
		return nil, errors.New("mux: route doesn't have a host")
	}
	values, err := r.prepareVars(pairs...)
	if err != nil {
		return nil, err
	}
	host, err := r.regexp.host.url(values)
	if err != nil {
		return nil, err
	}
	u := &url.URL{
		Scheme: "http",
		Host:   host,
	}
	if r.buildScheme != "" {
		u.Scheme = r.buildScheme
	}
	return u, nil
}

func (r *Route) URLPath(pairs ...string) (*url.URL, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.regexp.path == nil {
		return nil, errors.New("mux: route doesn't have a path")
	}
	values, err := r.prepareVars(pairs...)
	if err != nil {
		return nil, err
	}
	path, err := r.regexp.path.url(values)
	if err != nil {
		return nil, err
	}
	return &url.URL{
		Path: path,
	}, nil
}

func (r *Route) GetPathTemplate() (string, error) {
	if r.err != nil {
		return "", r.err
	}
	if r.regexp.path == nil {
		return "", errors.New("mux: route doesn't have a path")
	}
	return r.regexp.path.template, nil
}

func (r *Route) GetPathRegexp() (string, error) {
	if r.err != nil {
		return "", r.err
	}
	if r.regexp.path == nil {
		return "", errors.New("mux: route does not have a path")
	}
	return r.regexp.path.regexp.String(), nil
}

func (r *Route) GetQueriesRegexp() ([]string, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.regexp.queries == nil {
		return nil, errors.New("mux: route doesn't have queries")
	}
	queries := make([]string, 0, len(r.regexp.queries))
	for _, query := range r.regexp.queries {
		queries = append(queries, query.regexp.String())
	}
	return queries, nil
}

func (r *Route) GetQueriesTemplates() ([]string, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.regexp.queries == nil {
		return nil, errors.New("mux: route doesn't have queries")
	}
	queries := make([]string, 0, len(r.regexp.queries))
	for _, query := range r.regexp.queries {
		queries = append(queries, query.template)
	}
	return queries, nil
}

func (r *Route) GetMethods() ([]string, error) {
	if r.err != nil {
		return nil, r.err
	}
	for _, m := range r.matchers {
		if methods, ok := m.(methodMatcher); ok {
			return []string(methods), nil
		}
	}
	return nil, errors.New("mux: route doesn't have methods")
}

func (r *Route) GetHostTemplate() (string, error) {
	if r.err != nil {
		return "", r.err
	}
	if r.regexp.host == nil {
		return "", errors.New("mux: route doesn't have a host")
	}
	return r.regexp.host.template, nil
}

func (r *Route) prepareVars(pairs ...string) (map[string]string, error) {
	m, err := mapFromPairsToString(pairs...)
	if err != nil {
		return nil, err
	}
	return r.buildVars(m), nil
}

func (r *Route) buildVars(m map[string]string) map[string]string {
	if r.buildVarsFunc != nil {
		m = r.buildVarsFunc(m)
	}
	return m
}
