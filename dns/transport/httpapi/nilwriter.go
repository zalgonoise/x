package httpapi

import "net/http"

type responseWriter struct {
	header   int
	response string
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.response = string(b)
	return len(b), nil
}

func (w *responseWriter) WriteHeader(code int) {
	w.header = code
}

func (w *responseWriter) Header() http.Header {
	return nil
}
