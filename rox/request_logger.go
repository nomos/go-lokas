package rox

import (
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/util"
	"go.uber.org/zap"
	"net/http"
	"time"
)

var RequestLogger = CreateMiddleWare(func(w ResponseWriter, r *http.Request, a lokas.IProcess, next http.Handler) {
	start:=time.Now()
	log.Info("->"+r.Method,zap.String("path",r.Host+r.URL.Path),zap.Any("form",r.Form),zap.Int64("size",
		r.ContentLength),zap.String("ip",r.RemoteAddr),zap.String("time",util.FormatTimeToString(start)))
	next.ServeHTTP(w,r)
	resp:=w.(ResponseWriter)
	log.Info("<-"+r.Method,zap.String("path",r.Host+r.URL.Path),zap.Int("status",resp.Status()),zap.Any("resp",resp.GetResponse()),zap.Int("size",resp.Size()),zap.Int64("ms",time.Since(start).Milliseconds()))
})