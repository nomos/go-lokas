package util

import (
	"errors"
	"github.com/nomos/go-lokas/log"
	"runtime"
)

func Recover(r interface{},all bool)(err error){
	if err,ok:=r.(error);ok {
		log.Error(err.Error())
	}
	if str,ok:=r.(string);ok {
		log.Error(str)
		err = errors.New(str)
	}
	buf := make([]byte, 1<<14)
	runtime.Stack(buf, all)
	log.Error(string(buf))
	return
}
