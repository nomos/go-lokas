package rox

import (
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas"
	"net/http"
	"strconv"
	"strings"
)

type CorsOptions struct {
	AllowedOrigins []string

	AllowOriginFunc func(r *http.Request, origin string) bool

	AllowedMethods []string

	AllowedHeaders []string

	ExposedHeaders []string

	AllowCredentials bool

	MaxAge int

	OptionsPassthrough bool

	Debug bool
}

type Cors struct {
	allowedOrigins []string

	allowedWOrigins []wildcard

	allowOriginFunc func(r *http.Request, origin string) bool

	allowedHeaders []string

	allowedMethods []string

	exposedHeaders []string
	maxAge         int

	allowedOriginsAll bool

	allowedHeadersAll bool

	allowCredentials  bool
	optionPassthrough bool
	debug             bool
}

func NewCors(options CorsOptions) *Cors {
	c := &Cors{
		exposedHeaders:    convert(options.ExposedHeaders, http.CanonicalHeaderKey),
		allowOriginFunc:   options.AllowOriginFunc,
		allowCredentials:  options.AllowCredentials,
		maxAge:            options.MaxAge,
		optionPassthrough: options.OptionsPassthrough,
		debug:             options.Debug,
	}

	// Allowed Origins
	if len(options.AllowedOrigins) == 0 {
		if options.AllowOriginFunc == nil {
			// Default is all origins
			c.allowedOriginsAll = true
		}
	} else {
		c.allowedOrigins = []string{}
		c.allowedWOrigins = []wildcard{}
		for _, origin := range options.AllowedOrigins {
			// Normalize
			origin = strings.ToLower(origin)
			if origin == "*" {
				c.allowedOriginsAll = true
				c.allowedOrigins = nil
				c.allowedWOrigins = nil
				break
			} else if i := strings.IndexByte(origin, '*'); i >= 0 {
				w := wildcard{origin[0:i], origin[i+1:]}
				c.allowedWOrigins = append(c.allowedWOrigins, w)
			} else {
				c.allowedOrigins = append(c.allowedOrigins, origin)
			}
		}
	}

	if len(options.AllowedHeaders) == 0 {
		c.allowedHeaders = []string{"Origin", "Accept", "Content-Type"}
	} else {
		c.allowedHeaders = convert(append(options.AllowedHeaders, "Origin"), http.CanonicalHeaderKey)
		for _, h := range options.AllowedHeaders {
			if h == "*" {
				c.allowedHeadersAll = true
				c.allowedHeaders = nil
				break
			}
		}
	}

	if len(options.AllowedMethods) == 0 {
		c.allowedMethods = []string{http.MethodGet, http.MethodPost, http.MethodHead}
	} else {
		c.allowedMethods = convert(options.AllowedMethods, strings.ToUpper)
	}

	return c
}

func CorsAllowAll() *Cors {
	return NewCors(CorsOptions{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
	})
}

func (c *Cors) MiddleWare(w ResponseWriter, r *http.Request, a lokas.IProcess, next http.Handler) {
	//log.Warnf("Cors MiddleWare")
	if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
		c.logf("Handler: Preflight request")
		c.handlePreflight(w, r)
		if c.optionPassthrough {
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	} else {
		c.logf("Handler: Actual request")
		c.handleActualRequest(w, r)
		next.ServeHTTP(w, r)
	}
}

func (c *Cors) handlePreflight(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	origin := r.Header.Get("Origin")

	if r.Method != http.MethodOptions {
		c.logf("Preflight aborted: %s!=OPTIONS", r.Method)
		return
	}
	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Add("Vary", "Access-Control-Request-Headers")

	if origin == "" {
		c.logf("Preflight aborted: empty origin")
		return
	}
	if !c.isOriginAllowed(r, origin) {
		c.logf("Preflight aborted: origin '%s' not allowed", origin)
		return
	}

	reqMethod := r.Header.Get("Access-Control-Request-Method")
	if !c.isMethodAllowed(reqMethod) {
		c.logf("Preflight aborted: method '%s' not allowed", reqMethod)
		return
	}
	reqHeaders := parseHeaderList(r.Header.Get("Access-Control-Request-Headers"))
	if !c.areHeadersAllowed(reqHeaders) {
		c.logf("Preflight aborted: headers '%v' not allowed", reqHeaders)
		return
	}
	if c.allowedOriginsAll {
		headers.Set("Access-Control-Allow-Origin", "*")
	} else {
		headers.Set("Access-Control-Allow-Origin", origin)
	}

	headers.Set("Access-Control-Allow-Methods", strings.ToUpper(reqMethod))
	if len(reqHeaders) > 0 {

		headers.Set("Access-Control-Allow-Headers", strings.Join(reqHeaders, ", "))
	}
	if c.allowCredentials {
		headers.Set("Access-Control-Allow-Credentials", "true")
	}
	if c.maxAge > 0 {
		headers.Set("Access-Control-Max-Age", strconv.Itoa(c.maxAge))
	}
	c.logf("Preflight response headers: %v", headers)
}

func (c *Cors) handleActualRequest(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	origin := r.Header.Get("Origin")

	headers.Add("Vary", "Origin")
	if origin == "" {
		c.logf("Actual request no headers added: missing origin")
		return
	}
	if !c.isOriginAllowed(r, origin) {
		c.logf("Actual request no headers added: origin '%s' not allowed", origin)
		return
	}

	if !c.isMethodAllowed(r.Method) {
		c.logf("Actual request no headers added: method '%s' not allowed", r.Method)

		return
	}
	if c.allowedOriginsAll {
		headers.Set("Access-Control-Allow-Origin", "*")
	} else {
		headers.Set("Access-Control-Allow-Origin", origin)
	}
	if len(c.exposedHeaders) > 0 {
		headers.Set("Access-Control-Expose-Headers", strings.Join(c.exposedHeaders, ", "))
	}
	if c.allowCredentials {
		headers.Set("Access-Control-Allow-Credentials", "true")
	}
	c.logf("Actual response added headers: %v", headers)
}

func (c *Cors) logf(format string, a ...interface{}) {
	if c.debug {
		log.Warnf(concatInterface(format, a...))
	}
}

func (c *Cors) isOriginAllowed(r *http.Request, origin string) bool {
	if c.allowOriginFunc != nil {
		return c.allowOriginFunc(r, origin)
	}
	if c.allowedOriginsAll {
		return true
	}
	origin = strings.ToLower(origin)
	for _, o := range c.allowedOrigins {
		if o == origin {
			return true
		}
	}
	for _, w := range c.allowedWOrigins {
		if w.match(origin) {
			return true
		}
	}
	return false
}

func (c *Cors) isMethodAllowed(method string) bool {
	if len(c.allowedMethods) == 0 {
		// If no method allowed, always return false, even for preflight request
		return false
	}
	method = strings.ToUpper(method)
	if method == http.MethodOptions {
		// Always allow preflight requests
		return true
	}
	for _, m := range c.allowedMethods {
		if m == method {
			return true
		}
	}
	return false
}

func (c *Cors) areHeadersAllowed(requestedHeaders []string) bool {
	if c.allowedHeadersAll || len(requestedHeaders) == 0 {
		return true
	}
	for _, header := range requestedHeaders {
		header = http.CanonicalHeaderKey(header)
		found := false
		for _, h := range c.allowedHeaders {
			if h == header {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
