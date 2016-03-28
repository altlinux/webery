package main

import (
	"flag"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/altlinux/logfile-go"
	"github.com/altlinux/pidfile-go"
	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/ahttp/api"
	"github.com/altlinux/webery/pkg/ahttp/middleware/mlog"
	"github.com/altlinux/webery/pkg/config"
	"github.com/altlinux/webery/pkg/context"
	"github.com/altlinux/webery/pkg/db"
	storage "github.com/altlinux/webery/pkg/db/mongo"
)

var (
	configFile = flag.String("config", "", "Path to configuration file")
)

type Server struct {
	Cfg *config.Config
	DB  db.Session
}

func (s *Server) Handler(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = context.WithValue(ctx, db.ContextSession, s.DB)
	ctx = context.WithValue(ctx, config.ContextConfig, s.Cfg)
	ctx = context.WithValue(ctx, api.ContextEndpointsInfo, api.Endpoints)

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

	pidFile, err := pidfile.OpenPidfile(cfg.Global.Pidfile)
	if err != nil {
		log.Fatal("Unable to open pidfile: ", err.Error())
	}
	defer pidFile.Close()

	if err := pidFile.Check(); err != nil {
		log.Fatal("Check failed: ", err.Error())
	}

	if err := pidFile.Write(); err != nil {
		log.Fatal("Unable to write pidfile: ", err.Error())
	}

	logFile, err := logfile.OpenLogfile(cfg.Global.Logfile)
	if err != nil {
		log.Fatal("Unable to open log: ", err.Error())
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	dbi := storage.NewSession(cfg.Mongo)
	defer dbi.Close()

	server := Server{
		Cfg: cfg,
		DB:  dbi,
	}

	http.HandleFunc("/", server.Handler)
	log.Fatal(http.ListenAndServe(cfg.Global.Address, nil))
}
