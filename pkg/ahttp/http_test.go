package ahttp

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/altlinux/webery/pkg/ahttp/testresponse"
)

func TestHTTPResponse(t *testing.T) {
	p := testresponse.NewResponseWriter()
	w := NewResponseWriter(p)

	status := http.StatusInternalServerError
	errmsg := "FooBar"
	message := "Hello!"

	HTTPResponse(w, status, errmsg)
	w.Write([]byte("Hello!"))

	if w.(*ResponseWriter).HTTPStatus != status {
		t.Errorf("Unexpected HTTP status '%d', expected '%d'", w.(*ResponseWriter).HTTPStatus, status)
	}

	if w.(*ResponseWriter).HTTPError != errmsg {
		t.Errorf("Unexpected HTTP message '%s', expected '%s'", w.(*ResponseWriter).HTTPError, errmsg)
	}

	expect := fmt.Sprintf("Status: %d\n\n%s\n", status, message)
	result := p.(*testresponse.ResponseWriter).String()

	if expect != result {
		t.Errorf("Unexpected result")
	}
}
