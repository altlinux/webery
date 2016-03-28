package api

import (
	"net/http"
	"regexp"

	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/ahttp/middleware/db"
	"github.com/altlinux/webery/pkg/ahttp/middleware/jsonresponse"
	"github.com/altlinux/webery/pkg/context"
)

type apiEndpointsInfo int
type apiQueryParams int

const ContextEndpointsInfo apiEndpointsInfo = 0
const ContextQueryParams apiQueryParams = 0

type MethodHandlers map[string]ahttp.Handler

type HandlerInfo struct {
	Regexp          *regexp.Regexp
	Handlers        MethodHandlers
	NeedJSONHandler bool
	NeedDBHandler   bool
}

type EndpointsInfo struct {
	Endpoints []HandlerInfo
}

var Endpoints *EndpointsInfo = &EndpointsInfo{
	Endpoints: []HandlerInfo{
		{
			Regexp:          regexp.MustCompile("^/api/v1/search/?$"),
			NeedDBHandler:   false,
			NeedJSONHandler: true,
			Handlers: MethodHandlers{
				"GET": SearchHandler,
			},
		},
		{
			Regexp:          regexp.MustCompile("^/api/v1/tasks/(?P<task>[0-9]+)/subtasks/(?P<subtask>[0-9]+)/?$"),
			NeedDBHandler:   true,
			NeedJSONHandler: true,
			Handlers: MethodHandlers{
				"GET":    SubtaskGetHandler,
				"POST":   SubtaskUpdateHandler,
				"DELETE": SubtaskDeleteHandler,
			},
		},
		{
			Regexp:          regexp.MustCompile("^/api/v1/tasks/(?P<task>[0-9]+)/subtasks/?$"),
			NeedDBHandler:   true,
			NeedJSONHandler: true,
			Handlers: MethodHandlers{
				"GET":  SubtaskListHandler,
				"POST": SubtaskCreateHandler,
			},
		},
		{
			Regexp:          regexp.MustCompile("^/api/v1/tasks/(?P<task>[0-9]+)/?$"),
			NeedDBHandler:   true,
			NeedJSONHandler: true,
			Handlers: MethodHandlers{
				"GET":    TaskGetHandler,
				"POST":   TaskUpdateHandler,
				"DELETE": TaskDeleteHandler,
			},
		},
		{
			Regexp:          regexp.MustCompile("^/api/v1/tasks/?$"),
			NeedDBHandler:   true,
			NeedJSONHandler: true,
			Handlers: MethodHandlers{
				"GET":  TaskListHandler,
				"POST": TaskCreateHandler,
			},
		},
		{
			Regexp:          regexp.MustCompile("^/api/v1/acl/(?P<repo>[0-9a-z]+)/search/(?P<type>(groups|packages))/?$"),
			NeedDBHandler:   true,
			NeedJSONHandler: true,
			Handlers: MethodHandlers{
				"GET": AclListHandler,
			},
		},
		{
			Regexp:          regexp.MustCompile("^/api/v1/acl/(?P<repo>[0-9a-z]+)/(?P<type>(groups|packages))/(?P<name>[0-9A-Za-z_.-]+)/?$"),
			NeedDBHandler:   true,
			NeedJSONHandler: true,
			Handlers: MethodHandlers{
				"GET": AclGetHandler,
			},
		},
		{
			Regexp:          regexp.MustCompile("^/api/v1/acl/(?P<repo>[0-9a-z]+)/(?P<type>(groups|packages))/?$"),
			NeedDBHandler:   true,
			NeedJSONHandler: true,
			Handlers: MethodHandlers{
				"GET": AclListHandler,
			},
		},
		{
			Regexp:          regexp.MustCompile("^/api/v1/acl/?$"),
			NeedDBHandler:   false,
			NeedJSONHandler: true,
			Handlers: MethodHandlers{
				"GET": AclReposListHandler,
			},
		},
		{
			Regexp:          regexp.MustCompile("^/ping$"),
			NeedDBHandler:   false,
			NeedJSONHandler: false,
			Handlers: MethodHandlers{
				"GET": PingHandler,
			},
		},
		{
			Regexp:          regexp.MustCompile("^/"),
			NeedDBHandler:   false,
			NeedJSONHandler: false,
			Handlers: MethodHandlers{
				"GET": FileHandler,
			},
		},
	},
}

func Handler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	info, ok := ctx.Value(ContextEndpointsInfo).(*EndpointsInfo)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain API information from context")
		InternalServerErrorHandler(ctx, w, r)
		return
	}

	p := r.URL.Query()

	for _, a := range info.Endpoints {
		match := a.Regexp.FindStringSubmatch(r.URL.Path)
		if match == nil {
			continue
		}

		for i, name := range a.Regexp.SubexpNames() {
			if i == 0 {
				continue
			}
			p.Set(name, match[i])
		}

		ctx = context.WithValue(ctx, ContextQueryParams, &p)

		var reqHandler ahttp.Handler

		if v, ok := a.Handlers[r.Method]; ok {
			reqHandler = v

			if a.NeedDBHandler {
				reqHandler = db.Handler(reqHandler)
			}

			if a.NeedJSONHandler {
				reqHandler = jsonresponse.Handler(reqHandler)
			}
		} else {
			reqHandler = NotAllowedHandler
		}

		reqHandler(ctx, w, r)
		return
	}

	// Never should be here
	NotFoundHandler(ctx, w, r)
}
