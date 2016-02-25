package db

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/ahttp/acontext"
	"github.com/altlinux/webery/pkg/db"
)

type dbKeyRequestSession int

const ContextRequestSession dbKeyRequestSession = 0

func Handler(fn ahttp.Handler) ahttp.Handler {
	return func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
		dbi, ok := ctx.Value(db.ContextSession).(db.Session)

		if !ok {
			log.Fatalf("Unable to obtain database session from context")
			return
		}

		sess := dbi.Copy()
		ctx = acontext.WithValue(ctx, ContextRequestSession, sess)

		fn(ctx, resp, req)

		sess.Close()
	}
}
