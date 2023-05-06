package rox

import (
	"github.com/nomos/go-lokas/util"
	"reflect"
)
import "net/http"

// JSON response with optional status code.
func JSON(w ResponseWriter, val interface{}, code ...int) {
	if reflect.TypeOf(val).Kind() == reflect.Map {
		w.SetMap(util.MapToMap(val))
	} else {
		m, err := util.Struct2Map(val)
		w.SetMap(m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	if len(code) > 0 {
		w.WriteHeader(code[0])
	}
}
