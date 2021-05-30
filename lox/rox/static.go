package rox

import (
	"github.com/nomos/go-lokas"
	"net/http"
	"path"
	"strings"
)

type Static struct {
	Dir http.FileSystem

	Prefix string

	IndexFile string
}

func NewStatic(directory http.FileSystem) *Static {
	return &Static{
		Dir:       directory,
		Prefix:    "",
		IndexFile: "index.html",
	}
}

func (s *Static) MiddleWare(rw ResponseWriter, r *http.Request, a lokas.IProcess, next http.Handler) {
	if r.Method != "GET" && r.Method != "HEAD" {
		next.ServeHTTP(rw, r)
		return
	}
	file := r.URL.Path

	if s.Prefix != "" {
		if !strings.HasPrefix(file, s.Prefix) {
			next.ServeHTTP(rw, r)
			return
		}
		file = file[len(s.Prefix):]
		if file != "" && file[0] != '/' {
			next.ServeHTTP(rw, r)
			return
		}
	}
	f, err := s.Dir.Open(file)
	if err != nil {

		next.ServeHTTP(rw, r)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		next.ServeHTTP(rw, r)
		return
	}

	if fi.IsDir() {

		if !strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(rw, r, r.URL.Path+"/", http.StatusFound)
			return
		}

		file = path.Join(file, s.IndexFile)
		f, err = s.Dir.Open(file)
		if err != nil {
			next.ServeHTTP(rw, r)
			return
		}
		defer f.Close()

		fi, err = f.Stat()
		if err != nil || fi.IsDir() {
			next.ServeHTTP(rw, r)
			return
		}
	}

	http.ServeContent(rw, r, file, fi.ModTime(), f)
}
