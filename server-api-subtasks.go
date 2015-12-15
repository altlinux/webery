/*
 * Copyright (C) 2015 Alexey Gladkov <gladkov.alexey@gmail.com>
 *
 * This file is covered by the GNU General Public License,
 * which should be included with webery as the file COPYING.
 */

package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/altlinux/webery/misc"
	"github.com/altlinux/webery/model/subtask"
	"github.com/altlinux/webery/storage"
)

func (s *Server) apiListSubTaskHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	st := s.DB.NewStorage()
	defer st.Close()

	s.beginResponse(w, http.StatusOK)
	w.Write([]byte(`[`))

	iter := subtask.ListByTaskID(st, misc.ToInt64(p.Get("task"))).
		Sort("subtaskid").
		Iter()

	delim := false
	var subTask subtask.SubTask

	for iter.Next(&subTask) {
		msg, err := json.Marshal(subTask)
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

func (s *Server) apiGetSubTaskHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	st := s.DB.NewStorage()
	defer st.Close()

	taskID := misc.ToInt64(p.Get("task"))
	subtaskID := misc.ToInt64(p.Get("subtask"))

	res, err := subtask.GetSubTask(st, taskID, subtaskID)
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

func (s *Server) apiCreateSubTaskHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Unable to read body: %s", err)
		return
	}

	var data subtask.SubTask
	if err = json.Unmarshal(msg, &data); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Invalid JSON: %s", err)
		return
	}

	if data.SubTaskID == nil {
		s.errorResponse(w, http.StatusBadRequest, "subtaskid: mandatory field is not specified")
		return
	}

	if data.Owner == nil {
		s.errorResponse(w, http.StatusBadRequest, "owner: mandatory field is not specified")
		return
	}

	taskID := misc.ToInt64(p.Get("task"))
	data.TaskID = &taskID

	if data.Type == nil {
		s := "unknown"
		data.Type = &s
	}

	if data.Status == nil {
		s := "active"
		data.Status = &s
	}

	if err := subtask.Valid(data); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "%+v", err)
		return
	}

	st := s.DB.NewStorage()
	defer st.Close()

	if err := subtask.Create(st, data); err != nil {
		if storage.IsDup(err) {
			s.errorResponse(w, http.StatusBadRequest, "SubTask already exists")
		} else {
			s.errorResponse(w, httpStatusError(err), "%+v", err)
		}
		return
	}

	s.successResponse(w, "OK")

}

func (s *Server) apiUpdateSubTaskHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Unable to read body: %s", err)
		return
	}

	var data subtask.SubTask

	if err = json.Unmarshal(msg, &data); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Invalid JSON: %s", err)
		return
	}

	if err := subtask.Valid(data); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "%+v", err)
		return
	}

	st := s.DB.NewStorage()
	defer st.Close()

	taskID := misc.ToInt64(p.Get("task"))
	subtaskID := misc.ToInt64(p.Get("subtask"))

	if err := subtask.UpdateSubTask(st, taskID, subtaskID, data); err != nil {
		s.errorResponse(w, httpStatusError(err), "%+v", err)
		return
	}

	s.successResponse(w, "OK")
}

func (s *Server) apiDeleteSubTaskHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	st := s.DB.NewStorage()
	defer st.Close()

	taskID := misc.ToInt64(p.Get("task"))
	subtaskID := misc.ToInt64(p.Get("subtask"))

	if err := subtask.RemoveByID(st, taskID, subtaskID); err != nil {
		if err != storage.ErrNotFound {
			s.errorResponse(w, httpStatusError(err), "%+v", err)
		} else {
			s.notFoundHandler(w, r, p)
		}
		return
	}

	s.successResponse(w, "OK")
}
