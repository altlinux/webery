package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/ahttp/acontext"
	"github.com/altlinux/webery/pkg/ahttp/api"
	"github.com/altlinux/webery/pkg/ahttp/middleware/mlog"
	"github.com/altlinux/webery/pkg/config"
	"github.com/altlinux/webery/pkg/db"
	storage "github.com/altlinux/webery/pkg/db/mongo"
)

var (
	configFile = flag.String("config", "", "Path to configuration file")
	resultDoc  = "<html><head/><body><h1>Hello!</h1></body></html>"
)

func pageHandler(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
	time.Sleep(5 * time.Second)
	resp.WriteHeader(http.StatusOK)
	fmt.Fprintf(resp, "%s", resultDoc)
}

type Server struct {
	Cfg *config.Config
	DB  db.Session
}

func (s *Server) Handler(w http.ResponseWriter, req *http.Request) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	ctx = acontext.NewContext(ctx, req)
	ctx = acontext.WithValue(ctx, db.ContextSession, s.DB)
	ctx = acontext.WithValue(ctx, config.ContextConfig, s.Cfg)
	ctx = acontext.WithValue(ctx, api.ContextEndpointsInfo, api.Endpoints)

	mlog.Handler(api.Handler)(ctx, ahttp.NewResponseWriter(w), req)
}

func main() {
	flag.Parse()

	cfg, err := config.NewConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	log.SetLevel(cfg.Logging.Level.Level)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:    cfg.Logging.FullTimestamp,
		DisableTimestamp: cfg.Logging.DisableTimestamp,
		DisableColors:    cfg.Logging.DisableColors,
		DisableSorting:   cfg.Logging.DisableSorting,
	})

	dbi := storage.NewSession(cfg.Mongo)
	defer dbi.Close()

	server := Server{
		Cfg: cfg,
		DB:  dbi,
	}

	http.HandleFunc("/", server.Handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
