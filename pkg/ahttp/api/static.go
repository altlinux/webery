package api

import (
	"net/http"
	"os"

	"golang.org/x/net/context"

	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/config"
)

func FileHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if cfg, ok := ctx.Value(config.ContextConfig).(*config.Config); ok {
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
