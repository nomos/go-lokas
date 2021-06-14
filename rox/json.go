package rox

import (
	"github.com/nomos/go-lokas/util"
)
import "net/http"

// JSON response with optional status code.
func JSON(w ResponseWriter, val interface{}, code ...int) {
	var err error

	m,err:=util.Struct2Map(val)
	w.SetMap(m)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if len(code) > 0 {
		w.WriteHeader(code[0])
	}
}
