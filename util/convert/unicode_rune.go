package convert

import (
	"github.com/nomos/go-lokas/log"
)

func Unicode2Rune(s string)(string,error){
	rs,err:=Unicode2String(s)
	if err != nil {
		log.Error(err.Error())
		return "",err
	}
	return String2Rune(rs)
}

func Rune2Unicode(s string)(string,error){
	s1,err:=Rune2String(s)
	if err != nil {
		log.Error(err.Error())
		return "",err
	}
	return String2Unicode(s1)
}