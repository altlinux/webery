package mlog

import (
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/context"
)

func Handler(fn ahttp.Handler) ahttp.Handler {
	return func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
		w, ok := resp.(*ahttp.ResponseWriter)

		reqTime := time.Now()

		defer func() {
			e := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
				"timestop":  time.Now().String(),
				"timestart": reqTime.String(),
				"method":    req.Method,
				"client":    req.RemoteAddr,
				"reqlen":    req.ContentLength,
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
