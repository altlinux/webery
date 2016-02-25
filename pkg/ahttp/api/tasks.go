package api

import (
	//	"fmt"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"golang.org/x/net/context"

	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/db"
	"github.com/altlinux/webery/pkg/task"
	"github.com/altlinux/webery/pkg/util"
)

func TaskListHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	p, ok := ctx.Value(ContextQueryParams).(*url.Values)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain params from context")
		InternalServerErrorHandler(ctx, w, r)
		return
	}

	st, ok := ctx.Value(db.ContextSession).(db.Session)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain database from context")
		InternalServerErrorHandler(ctx, w, r)
		return
	}

	q := db.QueryDoc{}

	if p.Get("owner") != "" {
		q["owner"] = p.Get("owner")
	}
	if p.Get("repo") != "" {
		q["repo"] = p.Get("repo")
	}
	if p.Get("state") != "" {
		q["state"] = p.Get("state")
	}

	query, err := task.List(st, q)
	if err != nil {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to get tasks: %v", err)
		InternalServerErrorHandler(ctx, w, r)
		return
	}

	query = query.Sort("taskid")

	limit := util.ToInt32(p.Get("limit"))
	if limit > 0 {
		query = query.Limit(int(limit))
	}

	iter := query.Iter()
	successSent := false

	for {
		task := task.New()
		if !iter.Next(task) {
			break
		}

		if !ahttp.IsAlive(w) {
			return
		}

		if !successSent {
			successSent = true
			w.Write([]byte(`{`))
			w.Write([]byte(`"tasks":[`))
		} else {
			w.Write([]byte(`,`))
		}

		msg, err := json.Marshal(task)
		if err != nil {
			if !successSent {
				ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to marshal json: %v", err)
				InternalServerErrorHandler(ctx, w, r)
			}
			return
		}

		w.Write(msg)
	}
	if err := iter.Close(); err != nil {
		if !successSent {
			ahttp.HTTPResponse(w, http.StatusInternalServerError, "Error iterating tasks: %v", err)
			InternalServerErrorHandler(ctx, w, r)
		}
		return
	}

	if !successSent {
		w.Write([]byte(`{`))
		w.Write([]byte(`"tasks":[`))
	}
	w.Write([]byte(`]}`))
}

func TaskCreateHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	st, ok := ctx.Value(db.ContextSession).(db.Session)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain database from context")
		return
	}

	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "Unable to read body: %s", err)
		return
	}

	t := task.New()
	if err = json.Unmarshal(msg, t); err != nil {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "Invalid JSON: %s", err)
		return
	}

	if !t.TaskID.IsDefined() {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "taskid: mandatory field is not specified")
		return
	}

	if !t.Owner.IsDefined() {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "owner: mandatory field is not specified")
		return
	}

	if !t.Repo.IsDefined() {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "repo: mandatory field is not specified")
		return
	}

	if err := task.Write(st, t); err != nil {
		if db.IsDup(err) {
			ahttp.HTTPResponse(w, http.StatusBadRequest, "Task already exists")
		} else {
			ahttp.HTTPResponse(w, http.StatusInternalServerError, "%+v", err)
		}
		return
	}

	ahttp.HTTPResponse(w, http.StatusOK, "OK")
}
