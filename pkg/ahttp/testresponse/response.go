package testresponse

import (
	"bytes"
	"fmt"
	"net/http"
)

func NewResponseWriter() http.ResponseWriter {
	return &ResponseWriter{
		HTTPStatus:     http.StatusOK,
		ResponseHeader: make(http.Header),
	}
}

// ResponseWriter is a wrapper for http.ResponseWriter
type ResponseWriter struct {
	HTTPStatus     int
	HTTPError      string
	ResponseHeader http.Header
	ResponseData   bytes.Buffer
	ResponseLength int64
}

func (resp *ResponseWriter) Header() http.Header {
	return resp.ResponseHeader
}

func (resp *ResponseWriter) WriteHeader(status int) {
	resp.HTTPStatus = status
}

func (resp *ResponseWriter) Write(b []byte) (n int, err error) {
	n, err = resp.ResponseData.Write(b)
	if err == nil {
		resp.ResponseLength += int64(len(b))
	}
	return
}

func (resp *ResponseWriter) String() string {
	s := resp.ResponseData.String()
	return fmt.Sprintf("Status: %d\n\n%s\n", resp.HTTPStatus, s)
}

func (resp *ResponseWriter) StringHTTPValues() string {
	return fmt.Sprintf("Status: %d\nError: %s\n", resp.HTTPStatus, resp.HTTPError)
}
