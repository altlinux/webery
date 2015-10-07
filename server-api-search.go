/*
 * Copyright (C) 2015 Alexey Gladkov <gladkov.alexey@gmail.com>
 *
 * This file is covered by the GNU General Public License,
 * which should be included with webery as the file COPYING.
 */

package main

import (
	"net/http"
	"net/url"

	"gopkg.in/mgo.v2/bson"

	"github.com/altlinux/webery/misc"
)

var searchCollections []string = []string{"tasks", "subtasks"}

type searchResult struct {
	Query  string
	Result []interface{}
}

func (s *Server) apiSearchHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	prefix := p.Get("prefix")
	limit := misc.ToInt32(p.Get("limit"))

	if len(prefix) == 0 {
		s.beginResponse(w, http.StatusOK)
		w.Write([]byte(`[]`))
		s.endResponseSuccess(w)
		return
	}

	st := s.DB.NewStorage()
	defer st.Close()

	res := &searchResult{
		Query:  prefix,
		Result: make([]interface{}, 0),
	}

	num := int(limit)

	for _, name := range searchCollections {
		if limit > 0 && num == 0 {
			break
		}

		query := st.Coll(name).Find(bson.M{
			"search.key": bson.M{"$regex": "^"+prefix},
		})

		if num > 0 {
			query = query.Limit(num)
		}

		iter := query.Iter()

		var doc interface{}
		for iter.Next(&doc) {
			if !s.connIsAlive(w) {
				return
			}
			res.Result = append(res.Result, doc)
		}

		if err := iter.Close(); err != nil {
			s.errorResponse(w, httpStatusError(err), "error iterating: %+v", err)
			return
		}
	}

	s.successResponse(w, res)
}
