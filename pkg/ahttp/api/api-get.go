package api

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/context"

	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/db"
)

func apiGet(ctx context.Context, w http.ResponseWriter, r *http.Request, query Query) {
	st, ok := ctx.Value(db.ContextSession).(db.Session)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain database from context")
		return
	}

	col, err := st.Coll(query.CollName)
	if err != nil {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "%v", err)
		return
	}

	collQuery := col.Find(query.Pattern)
	doc, err := query.One(collQuery)

	if err != nil {
		if db.IsNotFound(err) {
			ahttp.HTTPResponse(w, http.StatusNotFound, "Not found")
		} else {
			ahttp.HTTPResponse(w, http.StatusInternalServerError, "%v", err)
		}
		return
	}

	msg, err := json.Marshal(doc)
	if err != nil {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to marshal document")
		return
	}

	w.Write([]byte(`{"result":`))
	w.Write(msg)
	w.Write([]byte(`}`))
}
