package db

import (
	"net/http"

	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/context"
	"github.com/altlinux/webery/pkg/db"
	"github.com/altlinux/webery/pkg/logger"
)

type dbKeyRequestSession int

const ContextRequestSession dbKeyRequestSession = 0

func Handler(fn ahttp.Handler) ahttp.Handler {
	return func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
		dbi, ok := ctx.Value("app.database").(db.Session)

		if !ok {
			logger.NewEntry().WithFields(nil).Fatalf("Unable to obtain database session from context")
			return
		}

		sess := dbi.Copy()
		ctx = context.WithValue(ctx, ContextRequestSession, sess)

		fn(ctx, resp, req)

		sess.Close()
	}
}
