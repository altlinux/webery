package api

import (
	//	"fmt"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/context"

	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/config"
	"github.com/altlinux/webery/pkg/db"
	"github.com/altlinux/webery/pkg/subtask"
	"github.com/altlinux/webery/pkg/task"
	"github.com/altlinux/webery/pkg/util"
)

func TaskListHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	p, ok := ctx.Value(ContextQueryParams).(*url.Values)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain params from context")
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

	apiSearch(ctx, w, r, []Query{
		Query{
			CollName: task.CollName,
			Sort:     []string{"taskid"},
			Pattern:  q,
			Iterator: func(iter db.Iter) interface{} {
				t := task.New()
				if !iter.Next(t) {
					return nil
				}
				return t
			},
		},
	})
}

func writeTask(ctx context.Context, w http.ResponseWriter, t *task.Task) bool {
	st, ok := ctx.Value(db.ContextSession).(db.Session)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain database from context")
		return false
	}

	cfg, ok := ctx.Value(config.ContextConfig).(*config.Config)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain config from context")
		return false
	}

	if v, ok := t.TaskID.Get(); ok {
		if v < int64(1) {
			ahttp.HTTPResponse(w, http.StatusBadRequest, "taskid must be greater than zero")
			return false
		}
	} else {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "taskid: mandatory field is not specified")
		return false
	}

	if v, ok := t.Repo.Get(); ok {
		if !util.InSliceString(v, cfg.Builder.Repos) {
			ahttp.HTTPResponse(w, http.StatusBadRequest, "Unknown repo")
			return false
		}
	} else {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "repo: mandatory field is not specified")
		return false
	}

	if !t.Owner.IsDefined() {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "owner: mandatory field is not specified")
		return false
	}

	if v, ok := t.State.Get(); ok {
		if !util.InSliceString(v, cfg.Builder.TaskStates) {
			ahttp.HTTPResponse(w, http.StatusBadRequest, "Unknown state")
			return false
		}
	}

	if err := task.Write(st, t); err != nil {
		if db.IsDup(err) {
			ahttp.HTTPResponse(w, http.StatusBadRequest, "Already exists")
		} else {
			ahttp.HTTPResponse(w, http.StatusInternalServerError, "%+v", err)
		}
		return false
	}

	return true
}

func TaskCreateHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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

	t.TimeCreate.Set(time.Now().Unix())

	if !writeTask(ctx, w, t) {
		return
	}

	ahttp.HTTPResponse(w, http.StatusOK, "OK")
}

func TaskGetHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	p, ok := ctx.Value(ContextQueryParams).(*url.Values)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain params from context")
		return
	}

	apiGet(ctx, w, r, Query{
		CollName: task.CollName,
		Pattern:  task.MakeID(util.ToInt64(p.Get("task"))),
		One: func(query db.Query) (interface{}, error) {
			t := task.New()
			err := query.One(t)
			return t, err
		},
	})
}

func TaskDeleteHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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

	taskID := task.MakeID(util.ToInt64(p.Get("task")))

	if err := task.Delete(st, taskID); err != nil {
		if db.IsNotFound(err) {
			ahttp.HTTPResponse(w, http.StatusNotFound, "Not found")
		} else {
			ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to delete: %v", err)
		}
		return
	}

	if err := subtask.Delete(st, taskID); err != nil {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to delete subtasks: %v", err)
		return
	}

	ahttp.HTTPResponse(w, http.StatusOK, "OK")
}

func TaskUpdateHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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

	taskID := task.MakeID(util.ToInt64(p.Get("task")))

	t, err := task.Read(st, taskID)
	if err != nil {
		if db.IsNotFound(err) {
			ahttp.HTTPResponse(w, http.StatusNotFound, "Not found")
		} else {
			ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to read: %v", err)
		}
		return
	}
	t.TimeCreate.Readonly(true)

	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "Unable to read body: %s", err)
		return
	}

	if err = json.Unmarshal(msg, t); err != nil {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "Invalid JSON: %s", err)
		return
	}

	if !writeTask(ctx, w, t) {
		return
	}

	ahttp.HTTPResponse(w, http.StatusOK, "OK")
}
