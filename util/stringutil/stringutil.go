package stringutil

import (
	"errors"
	"github.com/nomos/go-lokas/log"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

func Nil[T any]() T {
	var t T
	return t
}

func StartWithCapital(str string) bool {
	c1 := str[0]
	return c1 > 64 && c1 < 91
}

func SplitArray[T comparable](s string, split string) ([]T, error) {
	t := Nil[T]()
	arr := strings.Split(s, split)
	ret := []T{}
	for _, v := range arr {
		var v1 interface{} = v
		switch reflect.TypeOf(t).Kind() {
		case reflect.String:
			ret = append(ret, (v1).(T))
		case reflect.Float64, reflect.Float32:
			elem, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, err
			}
			ret = append(ret, interface{}(elem).(T))
		case reflect.Int8, reflect.Uint8, reflect.Int16, reflect.Uint16, reflect.Int32, reflect.Uint32, reflect.Int64, reflect.Uint64:
			intV, err := strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
			p := unsafe.Pointer(&intV)
			vp := (*T)(p)
			ret = append(ret, *vp)
		default:
			return nil, errors.New("unsupported type")
		}
	}
	return ret, nil
}

func SplitStringArray(s string, split string) []string {
	arr := strings.Split(s, split)
	return arr
}

func SplitFloat32Array(s string, split string) ([]float32, error) {
	arr := strings.Split(s, split)
	ret := make([]float32, 0)
	if s == "" {
		return ret, nil
	}
	for _, v := range arr {
		elem, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		ret = append(ret, float32(elem))
	}
	return ret, nil
}

func SplitFloat64Array(s string, split string) ([]float64, error) {
	arr := strings.Split(s, split)
	ret := make([]float64, 0)
	if s == "" {
		return ret, nil
	}
	for _, v := range arr {
		elem, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		ret = append(ret, float64(elem))
	}
	return ret, nil
}

func SplitInt32Array(s string, split string) ([]int32, error) {
	arr := strings.Split(s, split)
	ret := make([]int32, 0)
	if s == "" {
		return ret, nil
	}
	for _, v := range arr {
		elem, err := strconv.Atoi(v)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		ret = append(ret, int32(elem))
	}
	return ret, nil
}

func SplitInt32Map(s string, split string, mapSep string) (map[int32]int32, error) {
	arr := strings.Split(s, split)
	ret := make(map[int32]int32)
	if s == "" {
		return ret, nil
	}
	for _, v := range arr {
		elems := strings.Split(v, mapSep)
		if len(elems) != 2 {
			return nil, errors.New("SplitInt32Map:wrong format")
		}
		k, err := strconv.Atoi(elems[0])
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		v, err := strconv.Atoi(elems[1])
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		ret[int32(k)] = int32(v)
	}
	return ret, nil
}

func SplitInt64Array(s string, split string) ([]int64, error) {
	arr := strings.Split(s, split)
	ret := make([]int64, 0)
	if s == "" {
		return ret, nil
	}
	for _, v := range arr {
		elem, err := strconv.Atoi(v)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		ret = append(ret, int64(elem))
	}
	return ret, nil
}

func SplitIntArray(s string, split string) ([]int, error) {
	arr := strings.Split(s, split)
	ret := make([]int, 0)
	if s == "" {
		return ret, nil
	}
	for _, v := range arr {
		elem, err := strconv.Atoi(v)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		ret = append(ret, elem)
	}
	return ret, nil
}

func FirstToUpper(str string) string {
	var upperStr string
	vv := []rune(str) // 后文有介绍
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			if vv[i] >= 97 && vv[i] <= 122 { // 后文有介绍
				vv[i] -= 32 // string的码表相差32位
				upperStr += string(vv[i])
			} else {
				return str
			}
		} else {
			upperStr += string(vv[i])
		}
	}
	return upperStr
}

func CamelToSnake(str string) string {
	var upperStr string
	vv := []rune(str) // 后文有介绍
	for i := 0; i < len(vv); i++ {
		if vv[i] >= 65 && vv[i] <= 90 { // 后文有介绍
			if i != 0 {
				upperStr += "_"
			}
			vv[i] += 32 // string的码表相差32位
			upperStr += string(vv[i])
		} else {
			upperStr += string(vv[i])
		}
	}
	return upperStr
}

func FirstToLower(str string) string {
	var upperStr string
	vv := []rune(str) // 后文有介绍
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			if vv[i] >= 65 && vv[i] <= 90 { // 后文有介绍
				vv[i] += 32 // string的码表相差32位
				upperStr += string(vv[i])
			} else {
				return str
			}
		} else {
			upperStr += string(vv[i])
		}
	}
	return upperStr
}

func RandString(len int) string {
	if len <= 0 {
		len = 1
	}
	list := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	ret := ""
	for i := 0; i < len; i++ {
		if i == 0 {
			ret += string(list[rand.Intn(51)])
		} else {
			ret += string(list[rand.Intn(61)])
		}
	}
	log.Warnf("str", ret)
	return ret
}

func AddStringGap(str string, min int, gap int) string {
	delta2 := gap - len(str)%gap
	for {
		if len(str)+delta2 < min {
			delta2 += gap
			continue
		}
		break
	}
	for i := 0; i < delta2; i++ {
		str += " "
	}
	return str
}

func LocalPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Error(err.Error())
		return ""
	}
	return dir
}

func CopyString(str string) string {
	data := []byte(str)
	data1 := make([]byte, len(data))
	copy(data1, data)
	return string(data1)
}

func TrimEnd(str string) string {
	return regexp.MustCompile(`[\s*|\r|\n]*$`).ReplaceAllString(str, "")
}

func SplitCamelCase(src string) []string {
	// don't split invalid utf8
	if !utf8.ValidString(src) {
		return []string{src}
	}
	entries := []string{}
	var runes [][]rune
	lastClass := 0
	class := 0
	// split into fields based on class of unicode character
	for _, r := range src {
		switch true {
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			class = 3
		default:
			class = 4
		}
		if class == lastClass {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
		}
		lastClass = class
	}
	// handle upper case -> lower case sequences, e.g.
	// "PDFL", "oader" -> "PDF", "Loader"
	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}
	// construct []string from results
	for _, s := range runes {
		if len(s) > 0 {
			entries = append(entries, string(s))
		}
	}
	ret := []string{}
	for _, s := range entries {
		if regexp.MustCompile(`[_]*`).FindString(s) != s {
			ret = append(ret, s)
		}
	}
	return ret
}

func SplitCamelCaseUpperSlice(src string) (entries []string) {
	ret := SplitCamelCase(src)
	for i, v := range ret {
		ret[i] = strings.TrimSpace(strings.ToUpper(v))
	}
	return ret
}

func Capitalize(str string) string {
	var upperStr string
	vv := []rune(str) // 后文有介绍
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			if vv[i] >= 97 && vv[i] <= 122 { // 后文有介绍
				vv[i] -= 32 // string的码表相差32位
				upperStr += string(vv[i])
			} else {
				return str
			}
		} else {
			upperStr += string(vv[i])
		}
	}
	return upperStr
}

func SplitCamelCaseCapitalize(src string) (entries []string) {
	ret := SplitCamelCase(src)
	for i, v := range ret {
		ret[i] = strings.TrimSpace(Capitalize(v))
	}
	return ret
}

func SplitCamelCaseCapitalizeSlash(src string) string {
	ret := SplitCamelCaseCapitalize(src)
	return strings.Join(ret, "_")
}

func SnakeToCamel(src string) string {
	splits := strings.Split(src, "_")
	if len(splits) == 1 {
		return src
	}
	ret := ""
	for _, v := range splits {
		ret += Capitalize(v)
	}
	return ret
}

func SplitCamelCaseUpperSnake(src string) string {
	ret := SplitCamelCaseUpperSlice(src)
	return strings.Join(ret, "_")
}

func SplitCamelCaseLowerSlice(src string) (entries []string) {
	ret := SplitCamelCase(src)
	for i, v := range ret {
		ret[i] = strings.TrimSpace(strings.ToLower(v))
	}
	return ret
}

func SplitCamelCaseLowerSnake(src string) string {
	ret := SplitCamelCaseLowerSlice(src)
	return strings.Join(ret, "_")
}
