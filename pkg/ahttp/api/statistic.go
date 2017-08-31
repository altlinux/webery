package api

import (
	"encoding/json"
	"net/http"

	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/config"
	"github.com/altlinux/webery/pkg/context"
	"github.com/altlinux/webery/pkg/db"
	"github.com/altlinux/webery/pkg/logger"
	"github.com/altlinux/webery/pkg/task"
)

var taskStates = []string{"new", "awaiting", "postponed", "building", "pending", "committing"}

func StatisticQueueHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	st, ok := ctx.Value("app.database").(db.Session)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain database from context")
		return
	}

	cfg, ok := ctx.Value("app.config").(*config.Config)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain config from context")
		return
	}

	ans := make(map[string]map[string]int64)

	for _, repo := range cfg.Builder.Repos {
		ans[repo] = make(map[string]int64)
		for _, state := range taskStates {
			ans[repo][state] = 0
		}
	}

	q := db.QueryDoc{
		"state": db.QueryDoc{"$in": []string{"new", "awaiting", "postponed", "building", "pending", "committing"}},
	}

	col, err := st.Coll(task.CollName)
	if err != nil {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "%v", err)
		return
	}

	iter := col.Find(q).Iter()
	for {
		if !ahttp.IsAlive(w) {
			logger.GetHTTPEntry(ctx).WithFields(nil).Debugf("drop statistic answer because connection is not alive")
			return
		}

		t := task.New()
		if !iter.Next(t) {
			break
		}

		state, ok := t.State.Get()
		if !ok {
			continue
		}

		repo, ok := t.Repo.Get()
		if !ok {
			continue
		}

		ans[repo][state] += int64(1)
	}

	if err := iter.Close(); err != nil {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Error iterating: %+v", err)
		return
	}

	msg, err := json.Marshal(ans)
	if err != nil {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to marshal json: %v", err)
		return
	}

	w.Write([]byte(`{"result":`))
	w.Write(msg)
	w.Write([]byte(`}`))
}
