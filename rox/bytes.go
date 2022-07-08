package rox

import (
	"fmt"
	"net/http"
)

// bytes response with optional status code.
func Bytes(w ResponseWriter, fileName string, data []byte, code ...int) {
	var err error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", fileName))
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")
	w.WriteHeader(http.StatusOK)

	if len(code) > 0 {
		w.WriteHeader(code[0])
	}
}
