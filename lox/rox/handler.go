package rox

import (
	"github.com/nomos/go-lokas"
	"net/http"
)

type Handler func(w ResponseWriter,r *http.Request,a lokas.IProcess)



func CreateHandler(h Handler)Handler {
	return h
}