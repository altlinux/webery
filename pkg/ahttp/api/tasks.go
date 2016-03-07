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
			ahttp.HTTPResponse(w, http.StatusBadRequest, "Already exists")
		} else {
			ahttp.HTTPResponse(w, http.StatusInternalServerError, "%+v", err)
		}
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

	st, ok := ctx.Value(db.ContextSession).(db.Session)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain database from context")
		return
	}

	task, err := task.Read(st, task.MakeID(util.ToInt64(p.Get("task"))))
	if err != nil {
		if db.IsNotFound(err) {
			ahttp.HTTPResponse(w, http.StatusNotFound, "Not found")
		} else {
			ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to read: %v", err)
		}
		return
	}

	if !ahttp.IsAlive(w) {
		return
	}

	msg, err := json.Marshal(task)
	if err != nil {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "Unable to marshal json: %v", err)
		return
	}

	ignoreSubtasks := (p.Get("nosubtasks") != "")
	ignoreCancelled := (p.Get("nocancelled") != "")

	w.Write([]byte(`{`))
	w.Write([]byte(`"task":`))
	w.Write(msg)
	w.Write([]byte(`,"subtasks":[`))

	if !ignoreSubtasks {
		q := db.QueryDoc{
			"taskid": util.ToInt64(p.Get("task")),
		}

		query, err := subtask.List(st, q)
		if err != nil {
			ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to list: %v", err)
			return
		}

		iter := query.Sort("subtaskid").Iter()

		delim := false

		for {
			t := subtask.New()
			if !iter.Next(t) {
				break
			}

			if ignoreCancelled && t.IsCancelled() {
				continue
			}
			msg, err := json.Marshal(t)
			if err != nil {
				ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to marshal: %v", err)
				return
			}
			if delim {
				w.Write([]byte(`,`))
			}
			w.Write(msg)
			delim = true
		}

		if err := iter.Close(); err != nil {
			ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to close iterator: %v", err)
			return
		}
	}

	w.Write([]byte(`]}`))
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

	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "Unable to read body: %s", err)
		return
	}

	if err = json.Unmarshal(msg, t); err != nil {
		ahttp.HTTPResponse(w, http.StatusBadRequest, "Invalid JSON: %s", err)
		return
	}

	if err := task.Write(st, t); err != nil {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "%v", err)
		return
	}

	ahttp.HTTPResponse(w, http.StatusOK, "OK")
}
