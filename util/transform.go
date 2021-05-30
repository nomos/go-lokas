package util

import (
	"regexp"
)

func Binary2Byte(in string)uint8{
	if len(in)>8 {
		in = in[:8]
	}
	in = regexp.MustCompile(`[^0-1]`).ReplaceAllString(in,"0")
	var ret uint8 = 0
	for index,v:=range in {
		v2:=uint8(Ternary(v==48,0,1).(int))
		ret+=v2<<(7-index)
	}
	return ret
}


func Binary2Short(in string)uint16{
	if len(in)>16 {
		in = in[:16]
	}
	in = regexp.MustCompile(`[^0-1]`).ReplaceAllString(in,"0")
	var ret uint16 = 0
	for index,v:=range in {
		v2:=uint16(Ternary(v==48,0,1).(int))
		ret+=v2<<(15-index)
	}
	return ret
}