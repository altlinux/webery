package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"golang.org/x/net/context"
	"gopkg.in/mgo.v2/bson"

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

	st, ok := ctx.Value(db.ContextSession).(db.Session)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain database from context")
		return
	}

	res := &acl.ACL{}

	if p.Get("type") == "groups" {
		if p.Get("name") == "nobody" || p.Get("name") == "everybody" {
			res.Name = p.Get("name")
			res.Repo = p.Get("repo")
			res.Members = make([]acl.Member, 0)

			msg, err := json.Marshal(res)
			if err != nil {
				ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to marshal record")
				return
			}

			w.Write(msg)
			return
		}
	}

	coll, err := st.Coll("acl_" + p.Get("type"))
	if err != nil {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "%+v", err)
		return
	}

	query := coll.Find(bson.M{
		"repo": p.Get("repo"),
		"name": p.Get("name"),
	})

	if err := query.One(res); err != nil {
		if db.IsNotFound(err) {
			ahttp.HTTPResponse(w, http.StatusNotFound, "Not found")
		} else {
			ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to read: %v", err)
		}
		return
	}

	msg, err := json.Marshal(res)
	if err != nil {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to marshal record")
		return
	}

	w.Write(msg)
}
