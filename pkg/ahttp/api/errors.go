package api

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/ahttp/middleware/jsonresponse"
)

func JSONHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	resp, ok := w.(*ahttp.ResponseWriter)
	if !ok {
		panic("ResponseWriter is not ahttp.ResponseWriter")
	}

	jsonresponse.Handler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf(`{"status":%d,"title":"%s","detail":"%s"}`,
			resp.HTTPStatus,
			http.StatusText(resp.HTTPStatus),
			resp.HTTPError)))
	})(ctx, w, r)
}

func InternalServerErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	_, ok := w.(*ahttp.ResponseWriter)
	if !ok {
		w.Write([]byte(`Internal server error`))
		return
	}
	JSONHandler(ctx, w, r)
}

func NotFoundHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ahttp.HTTPResponse(w, http.StatusNotFound, "Page not found")
	JSONHandler(ctx, w, r)
}

func NotAllowedHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ahttp.HTTPResponse(w, http.StatusMethodNotAllowed, "Method Not Allowed")
	JSONHandler(ctx, w, r)
}

func PingHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ahttp.HTTPResponse(w, http.StatusOK, "OK")
	w.Write([]byte(""))
}
