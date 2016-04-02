package api

import (
	"net/http"
	"net/url"

	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/context"
	"github.com/altlinux/webery/pkg/db"
	"github.com/altlinux/webery/pkg/subtask"
	"github.com/altlinux/webery/pkg/task"
)

// :WEBAPI:
// {
//   "url": "{schema}://{host}/api/v1/search",
//   "method": "GET",
//   "parameters": [
//     {"name": "prefix", "type": "string", "description": "filter objects by prefix", "default": "NaN"},
//     {"name": "limit",  "type": "number", "description": "shows only specified number of retults", "default": "1000"}
//   ],
//   "description": "Returns a list of tasks and subtasks"
// }
func SearchHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	p, ok := ctx.Value("http.request.query.params").(*url.Values)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain params from context")
		return
	}

	apiSearch(ctx, w, r, []Query{
		Query{
			CollName: task.CollName,
			Pattern:  db.QueryDoc{"search.key": db.QueryDoc{"$regex": "^" + p.Get("prefix")}},
			Sort:     []string{"-taskid"},
			Iterator: func(iter db.Iter) interface{} {
				t := task.New()
				if !iter.Next(t) {
					return nil
				}
				return t
			},
		},
		Query{
			CollName: subtask.CollName,
			Pattern:  db.QueryDoc{"search.key": db.QueryDoc{"$regex": "^" + p.Get("prefix")}},
			Sort:     []string{"-taskid"},
			Iterator: func(iter db.Iter) interface{} {
				t := subtask.New()
				if !iter.Next(t) {
					return nil
				}
				return t
			},
		},
	})
}
