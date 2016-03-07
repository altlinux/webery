package api

import (
	"net/http"
	"net/url"

	"golang.org/x/net/context"

	"github.com/altlinux/webery/pkg/acl"
	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/db"
	"github.com/altlinux/webery/pkg/subtask"
	"github.com/altlinux/webery/pkg/task"
)

func SearchHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	p, ok := ctx.Value(ContextQueryParams).(*url.Values)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain params from context")
		return
	}

	apiSearch(ctx, w, r, []Query{
		Query{
			CollName: "tasks",
			Pattern:  db.QueryDoc{"search.key": db.QueryDoc{"$regex": "^" + p.Get("prefix")}},
			Iterator: func(iter db.Iter) interface{} {
				t := task.New()
				if !iter.Next(t) {
					return nil
				}
				return t
			},
		},
		Query{
			CollName: "subtasks",
			Pattern:  db.QueryDoc{"search.key": db.QueryDoc{"$regex": "^" + p.Get("prefix")}},
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

func AclSearchHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	p, ok := ctx.Value(ContextQueryParams).(*url.Values)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain params from context")
		return
	}

	query := db.QueryDoc{
		"repo": p.Get("repo"),
	}

	if p.Get("prefix") == "" {
		if p.Get("name") != "" {
			query["name"] = p.Get("name")
		}

		if p.Get("member") != "" {
			query["members.name"] = p.Get("member")
		}

	} else {
		query["name"] = db.QueryDoc{"$regex": "^" + p.Get("prefix")}
	}

	apiSearch(ctx, w, r, []Query{
		Query{
			CollName: "acl_" + p.Get("type"),
			Pattern:  query,
			Sort:     []string{"name"},
			Iterator: func(iter db.Iter) interface{} {
				t := &acl.ACL{}
				if !iter.Next(t) {
					return nil
				}
				return t
			},
		},
	})
}
