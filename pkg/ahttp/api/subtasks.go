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

func SubtaskCreateHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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

	t := subtask.New()
	if err = json.Unmarshal(msg, t); err != nil {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "Invalid JSON: %s", err)
		return
	}

	if !t.SubTaskID.IsDefined() {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "subtaskid: mandatory field is not specified")
		return
	}

	if !t.Owner.IsDefined() {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "owner: mandatory field is not specified")
		return
	}

	if !t.Type.IsDefined() {
		t.Type.Set("unknown")
	}

	if !t.Status.IsDefined() {
		t.Status.Set("active")
	}

	// TODO Validation

	if err := subtask.Write(st, t); err != nil {
		if db.IsDup(err) {
			ahttp.HTTPResponse(w, http.StatusBadRequest, "Already exists")
		} else {
			ahttp.HTTPResponse(w, http.StatusInternalServerError, "%+v", err)
		}
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

	msg, err := json.Marshal(t)
	if err != nil {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "Unable to marshal json: %v", err)
		return
	}

	w.Write(msg)
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

	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "Unable to read body: %s", err)
		return
	}

	if err = json.Unmarshal(msg, t); err != nil {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "Invalid JSON: %s", err)
		return
	}

	if err := subtask.Write(st, t); err != nil {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "%v", err)
		return
	}

	ahttp.HTTPResponse(w, http.StatusOK, "OK")
}
