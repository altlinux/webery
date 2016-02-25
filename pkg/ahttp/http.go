package ahttp

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"
)

type Handler func(context.Context, http.ResponseWriter, *http.Request)

func IsAlive(w http.ResponseWriter) bool {
	closeNotify := w.(http.CloseNotifier).CloseNotify()
	select {
	case closed := <-closeNotify:
		if closed {
			return false
		}
	default:
	}
	return true
}

func HTTPResponse(w http.ResponseWriter, status int, format string, args ...interface{}) {
	if resp, ok := w.(*ResponseWriter); ok {
		resp.HTTPStatus = status
		if format != "" {
			resp.HTTPError = fmt.Sprintf(format, args...)
		}
	}
	w.WriteHeader(status)
}
