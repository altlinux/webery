package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"golang.org/x/net/context"

	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/db"
	"github.com/altlinux/webery/pkg/util"
)

type Query struct {
	CollName string
	Pattern  db.QueryDoc
	Sort     []string
	Iterator func(db.Iter) interface{}
	One      func(db.Query) interface{}
}

func apiSearch(ctx context.Context, w http.ResponseWriter, r *http.Request, q []Query) {
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
