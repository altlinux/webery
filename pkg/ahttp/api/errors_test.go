package api

import (
	//	"fmt"
	"net/http"
	"testing"

	"golang.org/x/net/context"

	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/ahttp/acontext"
	"github.com/altlinux/webery/pkg/ahttp/testresponse"
)

func TestInternalServerErrorHandler(t *testing.T) {
	w := testresponse.NewResponseWriter()
	r := &http.Request{}

	ctx := acontext.NewContext(context.Background(), r)

	ahttp.HTTPResponse(w, http.StatusInternalServerError, "Error!")
	InternalServerErrorHandler(ctx, w, r)

	expect := "Status: 500\n\nInternal server error\n"
	result := w.(*testresponse.ResponseWriter).String()

	if expect != result {
		t.Errorf("Unexpected result: %q", result)
	}
}

func TestInternalServerErrorHandlerJSON(t *testing.T) {
	p := testresponse.NewResponseWriter()
	w := ahttp.NewResponseWriter(p)
	r := &http.Request{}

	ctx := acontext.NewContext(context.Background(), r)

	ahttp.HTTPResponse(w, http.StatusInternalServerError, "Error!")
	InternalServerErrorHandler(ctx, w, r)

	expect := "Status: 500\n\n{\"data\":{\"status\":500,\"title\":\"Internal Server Error\",\"detail\":\"Error!\"},\"status\":\"error\"}\n"
	result := p.(*testresponse.ResponseWriter).String()

	if expect != result {
		t.Errorf("Unexpected result: %q", result)
	}
}

func TestNotFoundHandler(t *testing.T) {
	p := testresponse.NewResponseWriter()
	w := ahttp.NewResponseWriter(p)
	r := &http.Request{}

	ctx := acontext.NewContext(context.Background(), r)

	NotFoundHandler(ctx, w, r)

	expect := "Status: 404\n\n{\"data\":{\"status\":404,\"title\":\"Not Found\",\"detail\":\"Page not found\"},\"status\":\"error\"}\n"
	result := p.(*testresponse.ResponseWriter).String()

	if expect != result {
		t.Errorf("Unexpected result: %q", result)
	}
}

func TestNotAllowedHandler(t *testing.T) {
	p := testresponse.NewResponseWriter()
	w := ahttp.NewResponseWriter(p)
	r := &http.Request{}

	ctx := acontext.NewContext(context.Background(), r)

	NotAllowedHandler(ctx, w, r)

	expect := "Status: 405\n\n{\"data\":{\"status\":405,\"title\":\"Method Not Allowed\",\"detail\":\"Method Not Allowed\"},\"status\":\"error\"}\n"
	result := p.(*testresponse.ResponseWriter).String()

	if expect != result {
		t.Errorf("Unexpected result: %q", result)
	}
}
