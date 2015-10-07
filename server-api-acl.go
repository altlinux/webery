/*
 * Copyright (C) 2015 Alexey Gladkov <gladkov.alexey@gmail.com>
 *
 * This file is covered by the GNU General Public License,
 * which should be included with webery as the file COPYING.
 */

package main

import (
	"encoding/json"
	"net/http"
	"net/url"

	"gopkg.in/mgo.v2/bson"

	"github.com/altlinux/webery/storage"
)

type ACLPerson struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Leader bool   `json:"leader"`
}

type ACL struct {
	Repo  string      `json:"repo"`
	Name  string      `json:"name"`
	Allow []ACLPerson `json:"allow"`
}

func (s *Server) apiListAclPackagesHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	st := s.DB.NewStorage()
	defer st.Close()

	s.beginResponse(w, http.StatusOK)
	w.Write([]byte(`[`))

	iter := st.Coll("acl_packages").
		Find(bson.M{"repo": p.Get("repo")}).
		Sort("name").
		Iter()

	delim := false
	var acl ACL

	for iter.Next(&acl) {
		if !s.connIsAlive(w) {
			return
		}

		msg, err := json.Marshal(acl)
		if err != nil {
			// FIXME legion: log failure
			return
		}
		if delim {
			w.Write([]byte(`,`))
		}
		w.Write(msg)
		delim = true
	}

	if err := iter.Close(); err != nil {
		// FIXME legion: log failure
		return
	}

	w.Write([]byte(`]`))
	s.endResponseSuccess(w)
}

func (s *Server) apiGetAclPackagesHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	st := s.DB.NewStorage()
	defer st.Close()

	query := st.Coll("acl_packages").
		Find(bson.M{
			"repo": p.Get("repo"),
			"name": p.Get("name"),
		})

	var res ACL
	if err := query.One(&res); err != nil {
		if err != storage.ErrNotFound {
			s.errorResponse(w, httpStatusError(err), "%+v", err)
		} else {
			s.notFoundHandler(w, r, p)
		}
		return
	}

	s.successResponse(w, res)
}

func (s *Server) apiListAclGroupsHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	st := s.DB.NewStorage()
	defer st.Close()

	s.beginResponse(w, http.StatusOK)
	w.Write([]byte(`[`))

	iter := st.Coll("acl_groups").
		Find(bson.M{"repo": p.Get("repo")}).
		Sort("name").
		Iter()

	delim := false
	var acl ACL

	for iter.Next(&acl) {
		if !s.connIsAlive(w) {
			return
		}

		msg, err := json.Marshal(acl)
		if err != nil {
			// FIXME legion: log failure
			return
		}
		if delim {
			w.Write([]byte(`,`))
		}
		w.Write(msg)
		delim = true
	}

	if err := iter.Close(); err != nil {
		// FIXME legion: log failure
		return
	}

	w.Write([]byte(`]`))
	s.endResponseSuccess(w)
}

func (s *Server) apiGetAclGroupsHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	var res ACL

	if p.Get("name") == "nobody" || p.Get("name") == "everybody" {
		res.Name = p.Get("name")
		res.Repo = p.Get("repo")
		res.Allow = make([]ACLPerson, 0)
		s.successResponse(w, res)
		return
	}

	st := s.DB.NewStorage()
	defer st.Close()

	query := st.Coll("acl_groups").
		Find(bson.M{
			"repo": p.Get("repo"),
			"name": p.Get("name"),
		})

	if err := query.One(&res); err != nil {
		if err != storage.ErrNotFound {
			s.errorResponse(w, httpStatusError(err), "%+v", err)
		} else {
			s.notFoundHandler(w, r, p)
		}
		return
	}

	s.successResponse(w, res)
}
