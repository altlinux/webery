package ahttp

import (
	"net/http"
	"sync"
)

func NewResponseWriter(w http.ResponseWriter) http.ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		HTTPStatus:     http.StatusOK,
	}
}

// ResponseWriter is a wrapper for http.ResponseWriter
type ResponseWriter struct {
	http.ResponseWriter

	mu sync.Mutex

	HTTPStatus     int
	HTTPError      string
	ResponseLength int64

	headerSent bool
}

func (resp *ResponseWriter) CloseNotify() <-chan bool {
	return resp.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (resp *ResponseWriter) WriteHeader(status int) {
	resp.mu.Lock()
	defer resp.mu.Unlock()

	resp.HTTPStatus = status
}

func (resp *ResponseWriter) Write(b []byte) (n int, err error) {
	resp.mu.Lock()
	defer resp.mu.Unlock()

	if !resp.headerSent {
		resp.ResponseWriter.WriteHeader(resp.HTTPStatus)
		resp.headerSent = true
	}
	n, err = resp.ResponseWriter.Write(b)
	if err == nil {
		resp.ResponseLength += int64(len(b))
	}
	return
}
