package util

import (
	"github.com/mitchellh/go-ps"
	"github.com/nomos/go-lokas/log"
)

func IsProcessExist(s string)bool{
	pses,err:=ps.Processes()
	if err != nil {
		log.Error(err.Error())
		return false
	}
	for _,p:=range pses {
		if p.Executable()==s {
			return true
		}
	}
	return false
}
