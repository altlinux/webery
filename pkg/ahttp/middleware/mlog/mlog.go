package mlog

import (
	"net/http"
	"time"

	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/context"
	"github.com/altlinux/webery/pkg/logger"
)

func Handler(fn ahttp.Handler) ahttp.Handler {
	return func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
		ctx = context.WithValue(ctx, "http.request.method", req.Method)
		ctx = context.WithValue(ctx, "http.request.remoteaddr", req.RemoteAddr)
		ctx = context.WithValue(ctx, "http.request.length", req.ContentLength)
		ctx = context.WithValue(ctx, "http.request.time", time.Now().String())

		defer func() {
			e := logger.GetHTTPEntry(ctx)
			e = e.WithField("http.response.time", time.Now().String())

			if w, ok := resp.(*ahttp.ResponseWriter); ok {
				e = e.WithField("http.response.length", w.ResponseLength)
				e = e.WithField("http.response.status", w.HTTPStatus)
				e = e.WithField("http.response.error", w.HTTPError)
			}
			e.Info(req.URL)
		}()

		fn(ctx, resp, req)
	}
}
