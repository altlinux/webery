package jsonresponse

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/altlinux/webery/pkg/ahttp"
)

func Handler(fn ahttp.Handler) ahttp.Handler {
	return func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "application/json")
		resp.Write([]byte(`{"data":`))

		resplen := int64(0)
		if w, ok := resp.(*ahttp.ResponseWriter); ok {
			resplen = w.ResponseLength
		}

		fn(ctx, resp, req)

		if w, ok := resp.(*ahttp.ResponseWriter); ok {
			if w.ResponseLength == resplen {
				w.Write([]byte(`{}`))
			}
			if w.HTTPStatus >= 200 && w.HTTPStatus < 300 {
				w.Write([]byte(`,"status":"success"`))
			} else {
				w.Write([]byte(`,"status":"error"`))
			}
		}

		resp.Write([]byte(`}`))
	}
}
