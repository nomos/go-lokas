package rox

import (
	"bufio"
	"encoding/json"
	"errors"
	"net"
	"net/http"

	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/protocol"
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

func NewResponseWriter(this http.ResponseWriter) ResponseWriter {
	nrw := &responseWriter{
		ResponseWriter: this,
		msg:            map[string]interface{}{},
		context:        map[string]interface{}{},
	}

	if _, ok := this.(http.CloseNotifier); ok {
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

func (this *responseWriter) WriteHeader(s int) {
	this.status = s
	this.callBefore()
	this.ResponseWriter.WriteHeader(s)
}

func (this *responseWriter) AddContext(k string, v interface{}) {
	this.context[k] = v
}

func (this *responseWriter) GetContext(k string) interface{} {
	return this.context[k]
}

func (this *responseWriter) SetText(msg string) {
	this.msg["text"] = msg
}

func (this *responseWriter) GetResponse() map[string]interface{} {
	return this.msg
}

func (this *responseWriter) SetMap(m map[string]interface{}) {
	for k, v := range m {
		this.msg[k] = v
	}
}

func (this *responseWriter) AddData(k string, v interface{}) {
	if this.msg["data"] == nil {
		this.msg["data"] = make(map[string]interface{})
	}
	this.msg["data"].(map[string]interface{})[k] = v
}

func (this *responseWriter) SetData(v interface{}) {
	this.msg["data"] = v
}

func (this *responseWriter) AddContent(k string, v interface{}) {
	this.msg[k] = v
}

func (this *responseWriter) WriteContent() error {
	if len(this.msg) == 0 {
		return nil
	}
	msg, err := json.Marshal(this.msg)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	_, err = this.Write(msg)
	if err != nil && err != http.ErrHijacked {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *responseWriter) Text() string {
	return this.msg["text"].(string)
}

func (this *responseWriter) Write(b []byte) (int, error) {
	if !this.Written() {

		this.WriteHeader(http.StatusOK)
	}
	size, err := this.ResponseWriter.Write(b)
	this.size += size
	return size, err
}

func (this *responseWriter) Status() int {
	return this.status
}

func (this *responseWriter) Size() int {
	return this.size
}

func (this *responseWriter) Written() bool {
	return this.status != 0
}

func (this *responseWriter) Before(before func(ResponseWriter)) {
	this.beforeFuncs = append(this.beforeFuncs, before)
}

func (this *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := this.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (this *responseWriter) callBefore() {
	for i := len(this.beforeFuncs) - 1; i >= 0; i-- {
		this.beforeFuncs[i](this)
	}
}

func (this *responseWriter) Flush() {
	flusher, ok := this.ResponseWriter.(http.Flusher)
	if ok {
		if !this.Written() {

			this.WriteHeader(http.StatusOK)
		}
		flusher.Flush()
	}
}

type responseWriterCloseNotifer struct {
	*responseWriter
}

func (this *responseWriterCloseNotifer) CloseNotify() <-chan bool {
	return this.ResponseWriter.(http.CloseNotifier).CloseNotify()
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
