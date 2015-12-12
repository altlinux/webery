/*
* Copyright (C) 2015 Alexey Gladkov <gladkov.alexey@gmail.com>
*
* This file is covered by the GNU General Public License,
* which should be included with webery as the file COPYING.
 */

package main

import (
	_ "net/http/pprof"

	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/altlinux/logfile-go"
	"github.com/altlinux/pidfile-go"

	"github.com/altlinux/webery/config"
	"github.com/altlinux/webery/logger"
	"github.com/altlinux/webery/storage"
)

var (
	configFile = flag.String("config", "", "Path to configuration file")
)

// JSONErrorData is a template for error answers.
type JSONErrorData struct {
	// HTTP status code.
	Code int `json:"status"`

	// Human readable error message.
	Message string `json:"detail"`
}

// HTTPResponse is a wrapper for http.ResponseWriter
type HTTPResponse struct {
	http.ResponseWriter

	HTTPStatus     int
	HTTPError      string
	ResponseLength int64
}

func (resp *HTTPResponse) Write(b []byte) (n int, err error) {
	n, err = resp.ResponseWriter.Write(b)
	if err == nil {
		resp.ResponseLength += int64(len(b))
	}
	return
}

// ConnTrack used to track the number of connections.
type ConnTrack struct {
	ConnID int64
	Conns  int64
}

type Server struct {
	lastConnID int64
	connsCount int64

	Cfg *config.Config
	Log *logger.ServerLogger
	DB  *storage.MongoService
}

func (s *Server) newConnTrack(r *http.Request) ConnTrack {
	cl := ConnTrack{
		ConnID: atomic.AddInt64(&s.lastConnID, 1),
	}

	conns := atomic.AddInt64(&s.connsCount, 1)
	log.Debugf("Opened connection %d (total=%d) [%s %s]", cl.ConnID, conns, r.Method, r.URL)

	cl.Conns = conns
	return cl
}

func (s *Server) closeConnTrack(cl ConnTrack) {
	conns := atomic.AddInt64(&s.connsCount, -1)
	log.Debugf("Closed connection %d (total=%d)", cl.ConnID, conns)
}

func (s *Server) connIsAlive(w *HTTPResponse) bool {
	closeNotify := w.ResponseWriter.(http.CloseNotifier).CloseNotify()

	select {
	case closed := <-closeNotify:
		if closed {
			return false
		}
	default:
	}
	return true
}

func (s *Server) writeStatus(resp *HTTPResponse, status int) {
	resp.HTTPStatus = status
	resp.WriteHeader(status)
}

func (s *Server) rawResponse(resp *HTTPResponse, status int, b []byte) {
	s.writeStatus(resp, status)
	resp.Write(b)
}

func (s *Server) beginResponse(w *HTTPResponse, status int) {
	//	s.Stats.HTTPStatus[status].Inc(1)

	w.Header().Set("Content-Type", "application/json")
	s.rawResponse(w, status, []byte(`{"data":`))
}

func (s *Server) endResponseSuccess(w *HTTPResponse) {
	w.Write([]byte(`,"status":"success"}`))
}

func (s *Server) beginResponseError(w *HTTPResponse, status int) {
	w.Header().Set("Content-Type", "application/json")
	s.rawResponse(w, status, []byte(`{"errors":`))
}

func (s *Server) endResponseError(w *HTTPResponse) {
	w.Write([]byte(`,"status":"error"}`))
}

func (s *Server) successResponse(w *HTTPResponse, m interface{}) {
	b, err := json.Marshal(m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorln("Unable to marshal result:", err)
		return
	}

	s.beginResponse(w, http.StatusOK)
	w.Write(b)
	s.endResponseSuccess(w)
}

func (s *Server) errorResponse(w *HTTPResponse, status int, format string, args ...interface{}) {
	w.HTTPError = fmt.Sprintf(format, args...)

	data := &JSONErrorData{
		Code:    status,
		Message: w.HTTPError,
	}
	log.Debugf("%+v", data)

	b, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorln("Unable to marshal result:", err)
		return
	}

	s.beginResponseError(w, status)
	w.Write([]byte(`[`))
	w.Write(b)
	w.Write([]byte(`]`))
	s.endResponseError(w)
}

func (s *Server) Run() error {
	type httpHandler struct {
		LimitConns    bool
		Regexp        *regexp.Regexp
		GETHandler    func(*HTTPResponse, *http.Request, *url.Values)
		POSTHandler   func(*HTTPResponse, *http.Request, *url.Values)
		DELETEHandler func(*HTTPResponse, *http.Request, *url.Values)
	}

	handlers := []httpHandler{
		httpHandler{
			Regexp:        regexp.MustCompile("^/api/v1/search/?$"),
			LimitConns:    true,
			GETHandler:    s.apiSearchHandler,
			POSTHandler:   s.notAllowedHandler,
			DELETEHandler: s.notAllowedHandler,
		},
		httpHandler{
			Regexp:        regexp.MustCompile("^/api/v1/tasks/(?P<task>[0-9]+)/subtasks/(?P<subtask>[0-9]+)/?$"),
			LimitConns:    true,
			GETHandler:    s.apiGetSubTaskHandler,
			POSTHandler:   s.apiUpdateSubTaskHandler,
			DELETEHandler: s.apiDeleteSubTaskHandler,
		},
		httpHandler{
			Regexp:        regexp.MustCompile("^/api/v1/tasks/(?P<task>[0-9]+)/subtasks/?$"),
			LimitConns:    true,
			GETHandler:    s.apiListSubTaskHandler,
			POSTHandler:   s.apiCreateSubTaskHandler,
			DELETEHandler: s.notAllowedHandler,
		},
		httpHandler{
			Regexp:        regexp.MustCompile("^/api/v1/tasks/(?P<task>[0-9]+)/?$"),
			LimitConns:    true,
			GETHandler:    s.apiGetTaskHandler,
			POSTHandler:   s.apiUpdateTaskHandler,
			DELETEHandler: s.apiDeleteTaskHandler,
		},
		httpHandler{
			Regexp:        regexp.MustCompile("^/api/v1/tasks/?$"),
			LimitConns:    true,
			GETHandler:    s.apiListTaskHandler,
			POSTHandler:   s.apiCreateTaskHandler,
			DELETEHandler: s.notAllowedHandler,
		},
		httpHandler{
			Regexp:        regexp.MustCompile("^/api/v1/acl/(?P<repo>[0-9a-z]+)/completion/(?P<type>(groups|packages))/?$"),
			LimitConns:    true,
			GETHandler:    s.apiAclCompletionHandler,
			POSTHandler:   s.notAllowedHandler,
			DELETEHandler: s.notAllowedHandler,
		},
		httpHandler{
			Regexp:        regexp.MustCompile("^/api/v1/acl/(?P<repo>[0-9a-z]+)/search/(?P<type>(groups|packages))/?$"),
			LimitConns:    true,
			GETHandler:    s.apiAclFindHandler,
			POSTHandler:   s.notAllowedHandler,
			DELETEHandler: s.notAllowedHandler,
		},
		httpHandler{
			Regexp:        regexp.MustCompile("^/api/v1/acl/(?P<repo>[0-9a-z]+)/packages/(?P<name>[0-9A-Za-z_.-]+)/?$"),
			LimitConns:    true,
			GETHandler:    s.apiGetAclPackagesHandler,
			POSTHandler:   s.notAllowedHandler,
			DELETEHandler: s.notAllowedHandler,
		},
		httpHandler{
			Regexp:        regexp.MustCompile("^/api/v1/acl/(?P<repo>[0-9a-z]+)/packages/?$"),
			LimitConns:    true,
			GETHandler:    s.apiListAclPackagesHandler,
			POSTHandler:   s.notAllowedHandler,
			DELETEHandler: s.notAllowedHandler,
		},
		httpHandler{
			Regexp:        regexp.MustCompile("^/api/v1/acl/(?P<repo>[0-9a-z]+)/groups/(?P<name>[0-9A-Za-z_.-]+)/?$"),
			LimitConns:    true,
			GETHandler:    s.apiGetAclGroupsHandler,
			POSTHandler:   s.notAllowedHandler,
			DELETEHandler: s.notAllowedHandler,
		},
		httpHandler{
			Regexp:        regexp.MustCompile("^/api/v1/acl/(?P<repo>[0-9a-z]+)/groups/?$"),
			LimitConns:    true,
			GETHandler:    s.apiListAclGroupsHandler,
			POSTHandler:   s.notAllowedHandler,
			DELETEHandler: s.notAllowedHandler,
		},
		httpHandler{
			Regexp:        regexp.MustCompile("^/api/v1/acl/?$"),
			LimitConns:    true,
			GETHandler:    s.apiListAclReposHandler,
			POSTHandler:   s.notAllowedHandler,
			DELETEHandler: s.notAllowedHandler,
		},
		httpHandler{
			Regexp:        regexp.MustCompile("^/(?P<path>[A-Za-z0-9/_-]+[.][a-z]+)$"),
			LimitConns:    true,
			GETHandler:    s.staticHandler,
			POSTHandler:   s.notAllowedHandler,
			DELETEHandler: s.notAllowedHandler,
		},
		httpHandler{
			Regexp:        regexp.MustCompile("^/ping$"),
			LimitConns:    false,
			GETHandler:    s.pingHandler,
			POSTHandler:   s.notAllowedHandler,
			DELETEHandler: s.notAllowedHandler,
		},
		httpHandler{
			Regexp:        regexp.MustCompile("^/"),
			LimitConns:    false,
			GETHandler:    s.rootHandler,
			POSTHandler:   s.notAllowedHandler,
			DELETEHandler: s.notAllowedHandler,
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/debug/vars", http.DefaultServeMux)
	mux.Handle("/debug/pprof/", http.DefaultServeMux)
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		reqTime := time.Now()
		resp := &HTTPResponse{w, http.StatusOK, "", 0}

		defer func() {
			e := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
				"stop":    time.Now().String(),
				"start":   reqTime.String(),
				"method":  req.Method,
				"addr":    req.RemoteAddr,
				"reqlen":  req.ContentLength,
				"resplen": resp.ResponseLength,
				"status":  resp.HTTPStatus,
			})

			if resp.HTTPStatus >= 500 {
				e = e.WithField("error", resp.HTTPError)
			}

			e.Info(req.URL)
		}()

		cl := s.newConnTrack(req)
		defer s.closeConnTrack(cl)

		p := req.URL.Query()

		for _, a := range handlers {
			match := a.Regexp.FindStringSubmatch(req.URL.Path)
			if match == nil {
				continue
			}

			if a.LimitConns && s.Cfg.Global.MaxConns > 0 && cl.Conns >= s.Cfg.Global.MaxConns {
				s.errorResponse(resp, http.StatusServiceUnavailable, "Too many connections")
				return
			}

			for i, name := range a.Regexp.SubexpNames() {
				if i == 0 {
					continue
				}
				p.Set(name, match[i])
			}

			switch req.Method {
			case "GET":
				a.GETHandler(resp, req, &p)
			case "POST":
				a.POSTHandler(resp, req, &p)
			case "DELETE":
				a.DELETEHandler(resp, req, &p)
			default:
				s.notAllowedHandler(resp, req, &p)
			}
			return
		}

		s.notFoundHandler(resp, req, &p)
		return
	})

	httpServer := &http.Server{
		Addr:    s.Cfg.Global.Address,
		Handler: mux,
	}

	log.Info("Server ready")
	return httpServer.ListenAndServe()
}

func main() {
	flag.Parse()

	if *configFile == "" {
		*configFile = "/etc/webery.cfg"
	}

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

	if cfg.Global.GoMaxProcs == 0 {
		cfg.Global.GoMaxProcs = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(cfg.Global.GoMaxProcs)

	server := Server{
		Cfg: cfg,
		Log: &logger.ServerLogger{
			Subsys: "server",
		},
		DB: storage.NewMongoService(cfg.Mongo),
	}

	st := server.DB.NewStorage()
	st.Initialize()
	st.Close()

	log.Fatal(server.Run())
}
