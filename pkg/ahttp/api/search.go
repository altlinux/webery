package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"golang.org/x/net/context"

	"github.com/altlinux/webery/pkg/acl"
	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/db"
	"github.com/altlinux/webery/pkg/subtask"
	"github.com/altlinux/webery/pkg/task"
	"github.com/altlinux/webery/pkg/util"
)

type Query struct {
	CollName string
	Pattern  db.QueryDoc
	Sort     []string
	Iterator func(db.Iter) interface{}
}

func Search(ctx context.Context, w http.ResponseWriter, r *http.Request, q []Query) {
	p, ok := ctx.Value(ContextQueryParams).(*url.Values)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain params from context")
		return
	}

	st, ok := ctx.Value(db.ContextSession).(db.Session)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain database from context")
		return
	}

	limit := util.ToInt32(p.Get("limit"))

	if limit == 0 {
		limit = 1000
	}

	delim := false
	w.Write([]byte(`{"result":[`))

	for _, query := range q {
		if !ahttp.IsAlive(w) {
			return
		}

		if limit == 0 {
			break
		}

		col, err := st.Coll(query.CollName)
		if err != nil {
			ahttp.HTTPResponse(w, http.StatusInternalServerError, "%v", err)
			return
		}

		collQuery := col.Find(query.Pattern).Limit(int(limit))

		if len(query.Sort) > 0 {
			collQuery = collQuery.Sort(query.Sort...)
		}

		iter := collQuery.Iter()

		for {
			if !ahttp.IsAlive(w) {
				return
			}

			if limit == 0 {
				break
			}

			doc := query.Iterator(iter)
			if doc == nil {
				break
			}

			msg, err := json.Marshal(doc)
			if err != nil {
				ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to marshal document")
				return
			}

			if delim {
				w.Write([]byte(`,`))
			}
			w.Write(msg)

			limit--
			delim = true
		}

		if err := iter.Close(); err != nil {
			ahttp.HTTPResponse(w, http.StatusInternalServerError, "Error iterating: %+v", err)
			return
		}
	}

	w.Write([]byte(`],"query":[`))

	delim = false
	for _, query := range q {
		msg, err := json.Marshal(query.Pattern)
		if err != nil {
			ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to marshal pattern")
			return
		}
		if delim {
			w.Write([]byte(`,`))
		}
		w.Write(msg)
		delim = true
	}

	w.Write([]byte(`]}`))
}

func SearchHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	p, ok := ctx.Value(ContextQueryParams).(*url.Values)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain params from context")
		return
	}

	Search(ctx, w, r, []Query{
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

	Search(ctx, w, r, []Query{
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
