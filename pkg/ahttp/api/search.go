package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"golang.org/x/net/context"
	"gopkg.in/mgo.v2/bson"

	"github.com/altlinux/webery/pkg/ahttp"
	"github.com/altlinux/webery/pkg/db"
	"github.com/altlinux/webery/pkg/util"
)

var searchCollections []string = []string{
	"tasks",
	"subtasks",
}

type searchResult struct {
	Query  string
	Result []interface{}
}

func SearchHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	p, ok := ctx.Value(ContextQueryParams).(*url.Values)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain params from context")
		InternalServerErrorHandler(ctx, w, r)
		return
	}

	prefix := p.Get("prefix")
	limit := util.ToInt32(p.Get("limit"))

	if len(prefix) == 0 {
		w.Write([]byte(`[]`))
		return
	}

	st, ok := ctx.Value(db.ContextSession).(db.Session)
	if !ok {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to obtain database from context")
		InternalServerErrorHandler(ctx, w, r)
		return
	}

	res := &searchResult{
		Query:  prefix,
		Result: make([]interface{}, 0),
	}

	num := int(limit)

	for _, name := range searchCollections {
		if limit > 0 && num == 0 {
			break
		}

		col, err := st.Coll(name)
		if err != nil {
		}

		query := col.Find(bson.M{
			"search.key": bson.M{
				"$regex": "^" + prefix,
			},
		})

		if num > 0 {
			query = query.Limit(num)
		}

		iter := query.Iter()

		var doc interface{}
		for iter.Next(&doc) {
			if !ahttp.IsAlive(w) {
				return
			}
			res.Result = append(res.Result, doc)
		}

		if err := iter.Close(); err != nil {
			ahttp.HTTPResponse(w, http.StatusInternalServerError, "Error iterating: %+v", err)
			InternalServerErrorHandler(ctx, w, r)
			return
		}
	}

	b, err := json.Marshal(res)
	if err != nil {
		ahttp.HTTPResponse(w, http.StatusInternalServerError, "Unable to marshal result: %v", err)
		InternalServerErrorHandler(ctx, w, r)
		return
	}

	w.Write(b)
}
