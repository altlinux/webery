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

	"gopkg.in/mgo.v2/bson"

	"github.com/altlinux/webery/misc"
	"github.com/altlinux/webery/model/subtask"
	"github.com/altlinux/webery/model/task"
	"github.com/altlinux/webery/storage"
)

func (s *Server) apiListTaskHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	st := s.DB.NewStorage()
	defer st.Close()

	q := bson.M{}

	if p.Get("owner") != "" {
		q["owner"] = p.Get("owner")
	}
	if p.Get("repo") != "" {
		q["repo"] = p.Get("repo")
	}
	if p.Get("state") != "" {
		q["state"] = p.Get("state")
	}

	query := task.List(st, q).Sort("taskid")

	limit := misc.ToInt32(p.Get("limit"))
	if limit > 0 {
		query = query.Limit(int(limit))
	}

	iter := query.Iter()

	successSent := false

	var task task.Task
	for iter.Next(&task) {
		if !s.connIsAlive(w) {
			return
		}

		if !successSent {
			successSent = true

			s.beginResponse(w, http.StatusOK)
			w.Write([]byte(`{`))
			w.Write([]byte(`"tasks":[`))
		} else {
			w.Write([]byte(`,`))
		}

		msg, err := json.Marshal(task)
		if err != nil {
			if !successSent {
				s.errorResponse(w, httpStatusError(err), "Unable to marshal json: %v", err)
			}
			// FIXME legion: log failure
			return
		}

		w.Write(msg)
	}
	if err := iter.Close(); err != nil {
		if !successSent {
			s.errorResponse(w, httpStatusError(err), "error iterating tasks: %+v", err)
		}
		// FIXME legion: log failure
		return
	}

	if !successSent {
		s.beginResponse(w, http.StatusOK)
		w.Write([]byte(`{`))
		w.Write([]byte(`"tasks":[`))
	}

	w.Write([]byte(`]}`))
	s.endResponseSuccess(w)
}

func (s *Server) apiGetTaskHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	st := s.DB.NewStorage()
	defer st.Close()

	task, err := task.GetTask(st, misc.ToInt64(p.Get("task")))
	if err != nil {
		if err != storage.ErrNotFound {
			s.errorResponse(w, httpStatusError(err), "%+v", err)
		} else {
			s.notFoundHandler(w, r, p)
		}
		return
	}

	if !s.connIsAlive(w) {
		return
	}

	msg, err := json.Marshal(task)
	if err != nil {
		s.errorResponse(w, httpStatusError(err), "Unable to marshal json: %v", err)
		return
	}

	ignoreSubtasks := (p.Get("nosubtasks") != "")
	ignoreCancelled := (p.Get("nocancelled") != "")

	s.beginResponse(w, http.StatusOK)
	w.Write([]byte(`{`))
	w.Write([]byte(`"task":`))
	w.Write(msg)

	if !ignoreSubtasks {
		w.Write([]byte(`,"subtasks":[`))

		delim := false

		iter := subtask.ListByTaskID(st, misc.ToInt64(p.Get("task"))).
			Sort("subtaskid").
			Iter()

		var subTask subtask.SubTask

		for iter.Next(&subTask) {
			if ignoreCancelled && subtask.IsSubTaskCancelled(subTask) {
				continue
			}
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
	}

	w.Write([]byte(`}`))
	s.endResponseSuccess(w)
}

func (s *Server) apiCreateTaskHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Unable to read body: %s", err)
		return
	}

	var t task.Task
	if err = json.Unmarshal(msg, &t); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Invalid JSON: %s", err)
		return
	}

	if err := task.Valid(t); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "%+v", err)
		return
	}

	st := s.DB.NewStorage()
	defer st.Close()

	if err := task.Create(st, t); err != nil {
		if storage.IsDup(err) {
			s.errorResponse(w, http.StatusBadRequest, "Task already exists")
		} else {
			s.errorResponse(w, httpStatusError(err), "%+v", err)
		}
		return
	}

	s.successResponse(w, "OK")
}

func (s *Server) apiUpdateTaskHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Unable to read body: %s", err)
		return
	}

	var data task.Task

	if err = json.Unmarshal(msg, &data); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Invalid JSON: %s", err)
		return
	}

	st := s.DB.NewStorage()
	defer st.Close()

	taskID := misc.ToInt64(p.Get("task"))

	if err := task.UpdateTask(st, taskID, data); err != nil {
		s.errorResponse(w, httpStatusError(err), "%+v", err)
		return
	}

	s.successResponse(w, "OK")
}

func (s *Server) apiDeleteTaskHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	st := s.DB.NewStorage()
	defer st.Close()

	taskID := misc.ToInt64(p.Get("task"))

	if err := task.RemoveByID(st, taskID); err != nil {
		if err != storage.ErrNotFound {
			s.errorResponse(w, httpStatusError(err), "%+v", err)
		} else {
			s.notFoundHandler(w, r, p)
		}
		return
	}

	if err := subtask.RemoveByTaskID(st, taskID); err != nil {
		s.errorResponse(w, httpStatusError(err), "%+v", err)
		return
	}

	s.successResponse(w, "OK")
}
