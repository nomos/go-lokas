package rox

import (
	"github.com/nomos/go-lokas/log"
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

func (this *Cors) MiddleWare(w ResponseWriter, r *http.Request, a lokas.IProcess, next http.Handler) {
	//log.Warnf("Cors MiddleWare")
	if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
		this.logf("Handler: Preflight request")
		this.handlePreflight(w, r)
		if this.optionPassthrough {
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	} else {
		this.logf("Handler: Actual request")
		this.handleActualRequest(w, r)
		next.ServeHTTP(w, r)
	}
}

func (this *Cors) handlePreflight(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	origin := r.Header.Get("Origin")

	if r.Method != http.MethodOptions {
		this.logf("Preflight aborted: %s!=OPTIONS", r.Method)
		return
	}
	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Add("Vary", "Access-Control-Request-Headers")

	if origin == "" {
		this.logf("Preflight aborted: empty origin")
		return
	}
	if !this.isOriginAllowed(r, origin) {
		this.logf("Preflight aborted: origin '%s' not allowed", origin)
		return
	}

	reqMethod := r.Header.Get("Access-Control-Request-Method")
	if !this.isMethodAllowed(reqMethod) {
		this.logf("Preflight aborted: method '%s' not allowed", reqMethod)
		return
	}
	reqHeaders := parseHeaderList(r.Header.Get("Access-Control-Request-Headers"))
	if !this.areHeadersAllowed(reqHeaders) {
		this.logf("Preflight aborted: headers '%v' not allowed", reqHeaders)
		return
	}
	if this.allowedOriginsAll {
		headers.Set("Access-Control-Allow-Origin", "*")
	} else {
		headers.Set("Access-Control-Allow-Origin", origin)
	}

	headers.Set("Access-Control-Allow-Methods", strings.ToUpper(reqMethod))
	if len(reqHeaders) > 0 {

		headers.Set("Access-Control-Allow-Headers", strings.Join(reqHeaders, ", "))
	}
	if this.allowCredentials {
		headers.Set("Access-Control-Allow-Credentials", "true")
	}
	if this.maxAge > 0 {
		headers.Set("Access-Control-Max-Age", strconv.Itoa(this.maxAge))
	}
	this.logf("Preflight response headers: %v", headers)
}

func (this *Cors) handleActualRequest(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	origin := r.Header.Get("Origin")

	headers.Add("Vary", "Origin")
	if origin == "" {
		this.logf("Actual request no headers added: missing origin")
		return
	}
	if !this.isOriginAllowed(r, origin) {
		this.logf("Actual request no headers added: origin '%s' not allowed", origin)
		return
	}

	if !this.isMethodAllowed(r.Method) {
		this.logf("Actual request no headers added: method '%s' not allowed", r.Method)

		return
	}
	if this.allowedOriginsAll {
		headers.Set("Access-Control-Allow-Origin", "*")
	} else {
		headers.Set("Access-Control-Allow-Origin", origin)
	}
	if len(this.exposedHeaders) > 0 {
		headers.Set("Access-Control-Expose-Headers", strings.Join(this.exposedHeaders, ", "))
	}
	if this.allowCredentials {
		headers.Set("Access-Control-Allow-Credentials", "true")
	}
	this.logf("Actual response added headers: %v", headers)
}

func (this *Cors) logf(format string, a ...interface{}) {
	if this.debug {
		log.Warnf(concatInterface(format, a...))
	}
}

func (this *Cors) isOriginAllowed(r *http.Request, origin string) bool {
	if this.allowOriginFunc != nil {
		return this.allowOriginFunc(r, origin)
	}
	if this.allowedOriginsAll {
		return true
	}
	origin = strings.ToLower(origin)
	for _, o := range this.allowedOrigins {
		if o == origin {
			return true
		}
	}
	for _, w := range this.allowedWOrigins {
		if w.match(origin) {
			return true
		}
	}
	return false
}

func (this *Cors) isMethodAllowed(method string) bool {
	if len(this.allowedMethods) == 0 {
		// If no method allowed, always return false, even for preflight request
		return false
	}
	method = strings.ToUpper(method)
	if method == http.MethodOptions {
		// Always allow preflight requests
		return true
	}
	for _, m := range this.allowedMethods {
		if m == method {
			return true
		}
	}
	return false
}

func (this *Cors) areHeadersAllowed(requestedHeaders []string) bool {
	if this.allowedHeadersAll || len(requestedHeaders) == 0 {
		return true
	}
	for _, header := range requestedHeaders {
		header = http.CanonicalHeaderKey(header)
		found := false
		for _, h := range this.allowedHeaders {
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
