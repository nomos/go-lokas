package util

import (
	"math"
	"strconv"
)

func DigitalNum(num int)int {
	for ret:=1;ret<100;ret++ {
		if num<int(math.Pow10(ret)) {
			return ret
		}
	}
	return 0
}

func DigitalToString(num int,digital int)string{
	d1:=DigitalNum(num)
	d2:=digital-d1
	ret:=strconv.Itoa(num)
	if d2>0 {
		for i:=0;i<d2;i++ {
			ret = "0"+ret
		}
	}
	return ret
}