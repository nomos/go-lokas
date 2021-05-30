package rox

import "strings"

const toLower = 'a' - 'A'

type converter func(string) string

type wildcard struct {
	prefix string
	suffix string
}

func (w wildcard) match(s string) bool {
	return len(s) >= len(w.prefix+w.suffix) && strings.HasPrefix(s, w.prefix) && strings.HasSuffix(s, w.suffix)
}

func convert(s []string, c converter) []string {
	out := []string{}
	for _, i := range s {
		out = append(out, c(i))
	}
	return out
}

func concatInterface(a interface{}, b ...interface{}) []interface{} {
	ret := make([]interface{}, 0)
	ret = append(ret, a)
	for _, v := range b {
		ret = append(ret, v)
	}
	return ret
}

func parseHeaderList(headerList string) []string {
	l := len(headerList)
	h := make([]byte, 0, l)
	upper := true

	t := 0
	for i := 0; i < l; i++ {
		if headerList[i] == ',' {
			t++
		}
	}
	headers := make([]string, 0, t)
	for i := 0; i < l; i++ {
		b := headerList[i]
		if b >= 'a' && b <= 'z' {
			if upper {
				h = append(h, b-toLower)
			} else {
				h = append(h, b)
			}
		} else if b >= 'A' && b <= 'Z' {
			if !upper {
				h = append(h, b+toLower)
			} else {
				h = append(h, b)
			}
		} else if b == '-' || (b >= '0' && b <= '9') {
			h = append(h, b)
		}

		if b == ' ' || b == ',' || i == l-1 {
			if len(h) > 0 {

				headers = append(headers, string(h))
				h = h[:0]
				upper = true
			}
		} else {
			upper = b == '-'
		}
	}
	return headers
}
