package convert

import (
	"fmt"
	"github.com/nomos/go-lokas/log"
	"strconv"
	"strings"
)

func String2Rune(s string)(string,error){
	runeDesc := ""
	for _, rune := range s {
		runeDesc += fmt.Sprintf("%d:", rune)
	}
	return runeDesc,nil
}

func Rune2String(s string)(string,error){
	splits:=strings.Split(s,":")
	runes:=[]rune{}
	for _,s:=range splits {
		r32,err:=strconv.Atoi(s)
		if err != nil {
			log.Error(err.Error())
			return "",err
		}
		r:=rune(int32(r32))
		runes = append(runes, r)
	}
	ret:=string(runes)
	return ret,nil
}