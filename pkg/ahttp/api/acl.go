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

	st, ok := ctx.Value(db.ContextSession).(db.Session)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain database from context")
		return
	}

	coll, err := st.Coll("acl_" + p.Get("type"))
	if err != nil {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "%+v", err)
		return
	}

	iter := coll.Find(bson.M{"repo": p.Get("repo")}).Sort("name").Iter()

	delim := false
	var rec acl.ACL

	w.Write([]byte(`[`))

	for iter.Next(&rec) {
		if !ahttp.IsAlive(w) {
			return
		}

		msg, err := json.Marshal(rec)
		if err != nil {
			ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to marshal record")
			return
		}
		if delim {
			w.Write([]byte(`,`))
		}
		w.Write(msg)
		delim = true
	}

	if err := iter.Close(); err != nil {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Error iterating: %v", err)
		return
	}

	w.Write([]byte(`]`))
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
