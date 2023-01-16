package http

import (
	"net/http"
	"strings"
)

func getToken(r *http.Request) (string, bool) {
	token := r.Header.Get("Authorization")
	if token != "" {
		t := strings.TrimPrefix(token, "Bearer ")
		if t != "" {
			return t, true
		}
	}
	return "", false
}

func getPath(path string) []string {
	splitPath := strings.Split(path, "/")
	var out = make([]string, 0, len(splitPath))
	for _, item := range splitPath {
		if item != "" && item != " " && item != "\n" && item != "\t" {
			out = append(out, item)
		}
	}
	return out
}
