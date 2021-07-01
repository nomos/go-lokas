package rox

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas"
	"net/http"
)

var FormParser = CreateMiddleWare(func(w ResponseWriter, r *http.Request, a lokas.IProcess, next http.Handler) {
	err:=r.ParseForm()
	if err != nil {
		log.Error(err.Error())
		http.Error(w,err.Error(),http.StatusBadRequest)
	}
	log.Warnf(r.Form)
	err=r.ParseMultipartForm(1024)
	if err != nil && err!= http.ErrNotMultipart{
		log.Error(err.Error())
		http.Error(w,err.Error(),http.StatusBadRequest)
	}
	if next!=nil {
		next.ServeHTTP(w,r)
	}
})