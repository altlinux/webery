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
	"github.com/altlinux/webery/pkg/util"
)

func SubtaskListHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	p, ok := ctx.Value(ContextQueryParams).(*url.Values)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain params from context")
		return
	}

	apiSearch(ctx, w, r, []Query{
		Query{
			CollName: subtask.CollName,
			Sort:     []string{"subtaskid"},
			Pattern:  db.QueryDoc{"taskid": util.ToInt64(p.Get("task"))},
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

func writeSubTask(ctx context.Context, w http.ResponseWriter, t *subtask.SubTask) bool {
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

	if v, ok := t.SubTaskID.Get(); ok {
		if v < int64(1) {
			ahttp.HTTPResponse(w, http.StatusBadRequest, "subtaskid must be greater than zero")
			return false
		}
	} else {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "subtaskid: mandatory field is not specified")
		return false
	}

	if v, ok := t.Type.Get(); ok {
		if !util.InSliceString(v, cfg.Builder.SubTaskTypes) {
			ahttp.HTTPResponse(w, http.StatusBadRequest, "Wrong value for 'type' field")
			return false
		}
	}

	if v, ok := t.Status.Get(); ok {
		if !util.InSliceString(v, cfg.Builder.SubTaskStates) {
			ahttp.HTTPResponse(w, http.StatusBadRequest, "Wrong value for 'status' field")
			return false
		}
	}

	if v, ok := t.CopyRepo.Get(); ok {
		if !util.InSliceString(v, cfg.Builder.Repos) {
			ahttp.HTTPResponse(w, http.StatusBadRequest, "Wrong value for 'copyrepo' field")
			return false
		}
	}

	if !t.Owner.IsDefined() {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "owner: mandatory field is not specified")
		return false
	}

	if !t.Type.IsDefined() {
		t.Type.Set("unknown")
	}

	if !t.Status.IsDefined() {
		t.Status.Set("active")
	}

	if err := subtask.Write(st, t); err != nil {
		if db.IsDup(err) {
			ahttp.HTTPResponse(w, http.StatusBadRequest, "Already exists")
		} else {
			ahttp.HTTPResponse(w, http.StatusInternalServerError, "%+v", err)
		}
		return false
	}
	return true
}

func SubtaskCreateHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "Unable to read body: %s", err)
		return
	}

	t := subtask.New()
	if err = json.Unmarshal(msg, t); err != nil {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "Invalid JSON: %s", err)
		return
	}

	t.TimeCreate.Set(time.Now().Unix())

	if !writeSubTask(ctx, w, t) {
		return
	}

	ahttp.HTTPResponse(w, http.StatusOK, "OK")
}

func SubtaskGetHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	p, ok := ctx.Value(ContextQueryParams).(*url.Values)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain params from context")
		return
	}

	apiGet(ctx, w, r, Query{
		CollName: subtask.CollName,
		Pattern:  subtask.MakeID(util.ToInt64(p.Get("task")), util.ToInt64(p.Get("subtask"))),
		One: func(query db.Query) (interface{}, error) {
			t := subtask.New()
			err := query.One(t)
			return t, err
		},
	})
}

func SubtaskDeleteHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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

	subtaskID := subtask.MakeID(util.ToInt64(p.Get("task")), util.ToInt64(p.Get("subtask")))

	if err := subtask.Delete(st, subtaskID); err != nil {
		if db.IsNotFound(err) {
			ahttp.HTTPResponse(w, http.StatusNotFound, "Not found")
		} else {
			ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to delete: %v", err)
		}
		return
	}

	ahttp.HTTPResponse(w, http.StatusOK, "OK")
}

func SubtaskUpdateHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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

	subtaskID := subtask.MakeID(util.ToInt64(p.Get("task")), util.ToInt64(p.Get("subtask")))

	t, err := subtask.Read(st, subtaskID)
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

	if !writeSubTask(ctx, w, t) {
		return
	}

	ahttp.HTTPResponse(w, http.StatusOK, "OK")
}
