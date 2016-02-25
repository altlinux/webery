package ahttp

import (
	"net/http"
)

func NewResponseWriter(w http.ResponseWriter) http.ResponseWriter {
	return &ResponseWriter{w, http.StatusOK, "", 0, false}
}

// ResponseWriter is a wrapper for http.ResponseWriter
type ResponseWriter struct {
	http.ResponseWriter

	HTTPStatus     int
	HTTPError      string
	ResponseLength int64

	headerSent bool
}

func (resp *ResponseWriter) CloseNotify() <-chan bool {
	return resp.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (resp *ResponseWriter) WriteHeader(status int) {
	resp.HTTPStatus = status
}

func (resp *ResponseWriter) Write(b []byte) (n int, err error) {
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
