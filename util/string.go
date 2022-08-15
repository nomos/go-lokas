package util

import (
	"fmt"
	"regexp"
	"strconv"
)

func ReplaceTextParams(text string, params ...interface{}) string {

	reg := regexp.MustCompile("#[0-9]+#")

	out := reg.ReplaceAllStringFunc(text, func(s string) string {

		idx, err := strconv.Atoi(s[1 : len(s)-1])
		if err != nil {
			return s
		}
		idx--
		if idx < 0 || idx > len(params)-1 {
			return s
		}

		switch params[idx].(type) {
		case string:
			return params[idx].(string)
		default:
			return fmt.Sprintf("%v", params[idx])
		}

	})

	return out
}
