package convert

import (
	"fmt"
	"strconv"
	"strings"
)

func Unicode2String(s string)(string,error){
	sUnicodev := strings.Split(s, "\\u")
	var context string
	for _, v := range sUnicodev {
		if len(v) < 1 {
			continue
		}
		temp, err := strconv.ParseInt(v, 16, 32)
		if err != nil {
			panic(err)
		}
		context += fmt.Sprintf("%c", temp)
	}
	return context,nil
}

func String2Unicode(s string)(string,error){
	textQuoted := strconv.QuoteToASCII(s)
	textUnquoted := textQuoted[1 : len(textQuoted)-1]
	return textUnquoted,nil
}
