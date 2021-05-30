package rox

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas/protocol"
	"net"
	"net/http"
)

type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher

	Status() int

	Written() bool

	Size() int

	Before(func(ResponseWriter))

	GetResponse() map[string]interface{}

	Text() string

	SetText(msg string)

	SetMap(m map[string]interface{})

	AddContent(k string, v interface{})

	AddData(k string, v interface{})

	AddContext(k string, v interface{})

	GetContext(k string) interface{}

	WriteContent() error

	Response(code int, msg ...interface{})

	OK(msg ...interface{})

	Failed(code protocol.ErrCode)

	JSON(val interface{}, code ...int)

	XML(val interface{}, code ...int)
}

type beforeFunc func(ResponseWriter)

func NewResponseWriter(rw http.ResponseWriter) ResponseWriter {
	nrw := &responseWriter{
		ResponseWriter: rw,
		msg:            map[string]interface{}{},
		context:        map[string]interface{}{},
	}

	if _, ok := rw.(http.CloseNotifier); ok {
		return &responseWriterCloseNotifer{nrw}
	}

	return nrw
}

type responseWriter struct {
	http.ResponseWriter
	status      int
	size        int
	msg         map[string]interface{}
	context     map[string]interface{}
	beforeFuncs []beforeFunc
}

func (rw *responseWriter) WriteHeader(s int) {
	rw.status = s
	rw.callBefore()
	rw.ResponseWriter.WriteHeader(s)
}

func (rw *responseWriter) AddContext(k string, v interface{}) {
	rw.context[k] = v
}

func (rw *responseWriter) GetContext(k string) interface{} {
	return rw.context[k]
}

func (rw *responseWriter) SetText(msg string) {
	rw.msg["text"] = msg
}

func (rw *responseWriter) GetResponse() map[string]interface{} {
	return rw.msg
}

func (rw *responseWriter) SetMap(m map[string]interface{}) {
	for k, v := range m {
		rw.msg[k] = v
	}
}

func (rw *responseWriter) AddData(k string, v interface{}) {
	if rw.msg["data"] == nil {
		rw.msg["data"] = make(map[string]interface{})
	}
	rw.msg["data"].(map[string]interface{})[k] = v
}

func (rw *responseWriter) SetData(v interface{}) {
	rw.msg["data"] = v
}

func (rw *responseWriter) AddContent(k string, v interface{}) {
	rw.msg[k] = v
}

func (rw *responseWriter) WriteContent() error {
	msg, err := json.Marshal(rw.msg)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	_, err = rw.Write(msg)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (rw *responseWriter) Text() string {
	return rw.msg["text"].(string)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.Written() {

		rw.WriteHeader(http.StatusOK)
	}
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) Size() int {
	return rw.size
}

func (rw *responseWriter) Written() bool {
	return rw.status != 0
}

func (rw *responseWriter) Before(before func(ResponseWriter)) {
	rw.beforeFuncs = append(rw.beforeFuncs, before)
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (rw *responseWriter) callBefore() {
	for i := len(rw.beforeFuncs) - 1; i >= 0; i-- {
		rw.beforeFuncs[i](rw)
	}
}

func (rw *responseWriter) Flush() {
	flusher, ok := rw.ResponseWriter.(http.Flusher)
	if ok {
		if !rw.Written() {

			rw.WriteHeader(http.StatusOK)
		}
		flusher.Flush()
	}
}

type responseWriterCloseNotifer struct {
	*responseWriter
}

func (rw *responseWriterCloseNotifer) CloseNotify() <-chan bool {
	return rw.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (w *responseWriter) Failed(code protocol.ErrCode) {
	arr := make([]interface{}, 0)
	arr = append(arr, code.Error())
	w.AddContent("errcode", code)
	write(w, http.StatusOK, arr)
}

func (w *responseWriter) Response(statusCode int, msg ...interface{}) {
	write(w, statusCode, msg)
}

func (w *responseWriter) OK(msg ...interface{}) {
	write(w, http.StatusOK, msg)
}

func (w *responseWriter) JSON(val interface{}, code ...int) {
	JSON(w, val, code...)
}

func (w *responseWriter) XML(val interface{}, code ...int) {
	XML(w, val, code...)
}
