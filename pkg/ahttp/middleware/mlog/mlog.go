package mlog

import (
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/altlinux/webery/pkg/ahttp"
)

func Handler(fn ahttp.Handler) ahttp.Handler {
	return func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
		w, ok := resp.(*ahttp.ResponseWriter)

		reqTime := time.Now()

		defer func() {
			e := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
				"stop":   time.Now().String(),
				"start":  reqTime.String(),
				"method": req.Method,
				"addr":   req.RemoteAddr,
				"reqlen": req.ContentLength,
			})

			if ok {
				e = e.WithField("resplen", w.ResponseLength)
				e = e.WithField("status", w.HTTPStatus)

				if w.HTTPStatus >= 500 {
					e = e.WithField("error", w.HTTPError)
				}
			}

			e.Info(req.URL)
		}()

		fn(ctx, resp, req)
	}
}
