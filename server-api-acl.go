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

	"github.com/altlinux/webery/config"
	"github.com/altlinux/webery/misc"
	"github.com/altlinux/webery/model/acl"
	"github.com/altlinux/webery/model/search"
	"github.com/altlinux/webery/storage"
)

func (s *Server) apiListAclReposHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	cfg := config.GetConfig()
	s.successResponse(w, cfg.Builder.Repos)
}

func (s *Server) apiListAclPackagesHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	st := s.DB.NewStorage()
	defer st.Close()

	s.beginResponse(w, http.StatusOK)
	w.Write([]byte(`[`))

	iter := acl.ListPackagesByRepo(st, p.Get("repo")).
		Sort("name").
		Iter()

	delim := false
	var rec acl.ACL

	for iter.Next(&rec) {
		if !s.connIsAlive(w) {
			return
		}

		msg, err := json.Marshal(rec)
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

	res, err := acl.GetPackageACL(st, p.Get("repo"), p.Get("name"))

	if err != nil {
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

	iter := acl.ListGroupsByRepo(st, p.Get("repo")).
		Sort("name").
		Iter()

	delim := false
	var rec acl.ACL

	for iter.Next(&rec) {
		if !s.connIsAlive(w) {
			return
		}

		msg, err := json.Marshal(rec)
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
	var res *acl.ACL

	if p.Get("name") == "nobody" || p.Get("name") == "everybody" {
		res = &acl.ACL{}
		res.Name = p.Get("name")
		res.Repo = p.Get("repo")
		res.Members = make([]acl.Member, 0)
		s.successResponse(w, res)
		return
	}

	st := s.DB.NewStorage()
	defer st.Close()

	res, err := acl.GetGroupACL(st, p.Get("repo"), p.Get("name"))

	if err != nil {
		if err != storage.ErrNotFound {
			s.errorResponse(w, httpStatusError(err), "%+v", err)
		} else {
			s.notFoundHandler(w, r, p)
		}
		return
	}

	s.successResponse(w, res)
}

func (s *Server) apiAclSearchHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	prefix := p.Get("prefix")
	repo := p.Get("repo")
	limit := misc.ToInt32(p.Get("limit"))

	if len(prefix) == 0 {
		s.beginResponse(w, http.StatusOK)
		w.Write([]byte(`[]`))
		s.endResponseSuccess(w)
		return
	}

	st := s.DB.NewStorage()
	defer st.Close()

	var doc acl.ACL

	num := int(limit)
	res := &searchResult{
		Query:  prefix,
		Result: make([]interface{}, 0),
	}

	iter := search.FindKey(st, "acl_packages", prefix, 0)
	for iter.Next(&doc) {
		if !s.connIsAlive(w) {
			return
		}
		if num == 0 {
			break
		}
		if doc.Repo != repo {
			continue
		}
		res.Result = append(res.Result, doc)
		num--
	}

	if err := iter.Close(); err != nil {
		s.errorResponse(w, httpStatusError(err), "error iterating: %+v", err)
		return
	}

	iter = search.FindPrefix(st, "acl_packages", prefix, 0)
	for iter.Next(&doc) {
		if !s.connIsAlive(w) {
			return
		}
		if num == 0 {
			break
		}
		if doc.Repo != repo || doc.Name == prefix {
			continue
		}
		res.Result = append(res.Result, doc)
		num--
	}

	if err := iter.Close(); err != nil {
		s.errorResponse(w, httpStatusError(err), "error iterating: %+v", err)
		return
	}

	s.successResponse(w, res)
}
