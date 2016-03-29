package api

import (
	"net/http"
	"os"

	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/config"
	"github.com/altlinux/webery/pkg/context"
)

func FileHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if cfg, ok := ctx.Value("app.config").(*config.Config); ok {
		path := r.URL.Path

		if _, err := os.Stat(cfg.Content.Path + path); err != nil {
			path = "/index.html"
		}

		http.ServeFile(w, r, cfg.Content.Path+path)
		return
	}
	ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain config from context")
	InternalServerErrorHandler(ctx, w, r)
}
