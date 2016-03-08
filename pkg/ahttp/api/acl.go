package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"golang.org/x/net/context"

	"github.com/altlinux/webery/pkg/acl"
	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/config"
	"github.com/altlinux/webery/pkg/db"
)

func AclReposListHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if cfg, ok := ctx.Value(config.ContextConfig).(*config.Config); ok {
		msg, err := json.Marshal(cfg.Builder.Repos)
		if err != nil {
			ahttp.HTTPResponse(w, http.StatusBadRequest, "Unable to marshal json: %v", err)
			return
		}
		w.Write(msg)
		return
	}
	ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain config from context")
}

func AclListHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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
			Sort:     []string{"name"},
			Pattern:  query,
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

func AclGetHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	p, ok := ctx.Value(ContextQueryParams).(*url.Values)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain params from context")
		return
	}

	apiGet(ctx, w, r, Query{
		CollName: "acl_" + p.Get("type"),
		Pattern:  db.QueryDoc{
			"repo": p.Get("repo"),
			"name": p.Get("name"),
		},
		One:      func(query db.Query) (interface{}, error) {
			var err error
			t := &acl.ACL{}

			if p.Get("type") == "groups" {
				if p.Get("name") == "nobody" || p.Get("name") == "everybody" {
					t.Name = p.Get("name")
					t.Repo = p.Get("repo")
					t.Members = make([]acl.Member, 0)
					return t, err
				}
			}

			err = query.One(t)
			return t, err
		},
	})
}
