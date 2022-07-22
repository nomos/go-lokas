package util

import (
	"bytes"
	"errors"
	"github.com/nomos/go-lokas/log"
	"runtime"
)

func Recover(r interface{}, all bool) error {
	var err error
	var ok bool
	if err, ok = r.(error); ok {
		log.Error(err.Error())
	}
	var str string
	if str, ok = r.(string); ok {
		log.Error(str)
		err = errors.New(str)
	}
	buf := make([]byte, 1<<14)
	runtime.Stack(buf, all)
	buf = bytes.TrimRightFunc(buf, func(r rune) bool {
		return r == 0
	})
	log.Error(string(buf))
	return err
}
