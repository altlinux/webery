package ahttp

import (
	"fmt"
	"net/http"

	"github.com/altlinux/webery/pkg/context"
	"github.com/altlinux/webery/pkg/logger"
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
	err := ""
	if format != "" {
		err = fmt.Sprintf(format, args...)
	}

	if resp, ok := w.(*ResponseWriter); ok {
		resp.HTTPStatus = status
		resp.HTTPError = err
	}

	if 400 <= status {
		logger.WithFieldsDepth(nil, 3).Error(err)
	}
	w.WriteHeader(status)
}
