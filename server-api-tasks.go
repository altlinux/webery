/*
 * Copyright (C) 2015 Alexey Gladkov <gladkov.alexey@gmail.com>
 *
 * This file is covered by the GNU General Public License,
 * which should be included with webery as the file COPYING.
 */

package main

import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/altlinux/webery/misc"
	"github.com/altlinux/webery/storage"
)

type Keyword struct {
	Key string
}

type SubTask struct {
	ObjType   string
	TaskID    int64
	SubTaskID int64
	Owner     string
	Type      string
	Status    string
	PkgName   string
	TagName   string
	TagID     string
	TagAuthor string
	Dir       string
	CopyRepo  string
	TimeCreate int64
	Search    []Keyword
}

type Task struct {
	ObjType    string
	TaskID     int64
	Try        int64
	Iter       int64
	Shared     bool
	TestOnly   bool
	State      string
	Repo       string
	Owner      string
	TimeCreate int64
	Search     []Keyword
}

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

	query := st.Coll("tasks").
		Find(q).
		Sort("taskid")

	limit := misc.ToInt32(p.Get("limit"))
	if limit > 0 {
		query = query.Limit(int(limit))
	}

	iter := query.Iter()

	successSent := false

	var task Task
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
	var task Task

	st := s.DB.NewStorage()
	defer st.Close()

	q := st.Coll("tasks").
		Find(bson.M{
			"taskid": misc.ToInt64(p.Get("task")),
		})

	if err := q.One(&task); err != nil {
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

	ignoreSubtasks  := (p.Get("nosubtasks") != "")
	ignoreCancelled := (p.Get("nocancelled") != "")

	s.beginResponse(w, http.StatusOK)
	w.Write([]byte(`{`))
	w.Write([]byte(`"task":`))
	w.Write(msg)

	if !ignoreSubtasks {
		w.Write([]byte(`,"subtasks":[`))

		delim := false

		iter := st.Coll("subtasks").
			Find(bson.M{
				"taskid": misc.ToInt64(p.Get("task")),
			}).
			Sort("subtaskid").
			Iter()

		var subTask SubTask
		for iter.Next(&subTask) {
			if ignoreCancelled && subTask.Status == "cancelled" {
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

	var task Task
	if err = json.Unmarshal(msg, &task); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Invalid JSON: %s", err)
		return
	}

	if task.TaskID <= 0 {
		s.errorResponse(w, http.StatusBadRequest, "Bad TaskID")
		return
	}

	task.Repo = strings.ToLower(task.Repo)

	if !misc.InSliceString(task.Repo, s.Cfg.Builder.Repos) {
		s.errorResponse(w, http.StatusBadRequest, "Unknown repo: %s", task.Repo)
		return
	}

	task.ObjType = "task"
	task.TimeCreate = time.Now().Unix()

	task.Search = make([]Keyword, 0)
	task.Search = append(task.Search, Keyword{fmt.Sprint(task.TaskID)})
	task.Search = append(task.Search, Keyword{task.Repo})

	st := s.DB.NewStorage()
	defer st.Close()

	if err := st.Coll("tasks").Insert(&task); err != nil {
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

	st := s.DB.NewStorage()
	defer st.Close()

	type updateTask struct {
		TaskID   int64
		Shared   *bool
		TestOnly *bool
		Try      *int64
		Iter     *int64
		State    *string
	}

	var data updateTask
	if err = json.Unmarshal(msg, &data); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Invalid JSON: %s", err)
		return
	}

	changeSet := bson.M{}

	if data.Shared != nil {
		changeSet["shared"] = *data.Shared
	}

	if data.TestOnly != nil {
		changeSet["testonly"] = *data.TestOnly
	}

	if data.State != nil {
		state := strings.ToLower(*data.State)

		if !misc.InSliceString(state, s.Cfg.Builder.TaskStates) {
			s.errorResponse(w, http.StatusBadRequest, "Wrong value for 'status' field")
			return
		}
		changeSet["status"] = state
	}

	changeInc := bson.M{}

	if data.Try != nil {
		if *data.Try <= 0 {
			s.errorResponse(w, http.StatusBadRequest, "'try' must be greater than zero")
			return
		} else {
			changeInc["try"] = *data.Try
		}
	}

	if data.Iter != nil {
		if *data.Iter <= 0 {
			s.errorResponse(w, http.StatusBadRequest, "'iter' must be greater than zero")
			return
		} else {
			changeInc["iter"] = *data.Iter
		}
	}

	change := bson.M{}

	if len(changeSet) > 0 {
		change["$set"] = changeSet
	}
	if len(changeInc) > 0 {
		change["$inc"] = changeInc
	}

	if len(change) == 0 {
		s.errorResponse(w, http.StatusBadRequest, "More fields required")
		return
	}

	err = st.Coll("tasks").Update(bson.M{"taskid": misc.ToInt64(p.Get("task"))}, change)
	if err != nil {
		s.errorResponse(w, httpStatusError(err), "%+v", err)
		return
	}

	s.successResponse(w, "OK")
}

func (s *Server) apiDeleteTaskHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	st := s.DB.NewStorage()
	defer st.Close()

	query := bson.M{
		"taskid": misc.ToInt64(p.Get("task")),
	}

	if err := st.Coll("tasks").Remove(query); err != nil {
		if err != storage.ErrNotFound {
			s.errorResponse(w, httpStatusError(err), "%+v", err)
		} else {
			s.notFoundHandler(w, r, p)
		}
		return
	}

	s.successResponse(w, "OK")
}

func (s *Server) apiListSubtaskHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	st := s.DB.NewStorage()
	defer st.Close()

	s.beginResponse(w, http.StatusOK)
	w.Write([]byte(`[`))

	iter := st.Coll("subtasks").
		Find(bson.M{
			"taskid": misc.ToInt64(p.Get("task")),
		}).
		Sort("subtaskid").
		Iter()

	delim := false
	var subTask SubTask

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

func (s *Server) apiGetSubtaskHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	st := s.DB.NewStorage()
	defer st.Close()

	query := st.Coll("subtasks").
		Find(bson.M{
			"taskid":    misc.ToInt64(p.Get("task")),
			"subtaskid": misc.ToInt64(p.Get("subtask")),
		})

	var res SubTask

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

func (s *Server) apiCreateSubTaskHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Unable to read body: %s", err)
		return
	}

	var data SubTask
	if err = json.Unmarshal(msg, &data); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Invalid JSON: %s", err)
		return
	}

	if data.SubTaskID <= 0 {
		s.errorResponse(w, http.StatusBadRequest, "Bad SubTaskID")
		return
	}

	data.ObjType = "subtask"
	data.TaskID = misc.ToInt64(p.Get("task"))

	if data.TaskID <= 0 {
		s.errorResponse(w, http.StatusBadRequest, "Bad TaskID")
		return
	}

	data.TimeCreate = time.Now().Unix()
	data.Search = make([]Keyword, 0)

	if data.Status != "cancelled" {
		if len(data.Owner) > 0 {
			data.Search = append(data.Search, Keyword{data.Owner})
		}

		if len(data.PkgName) > 0 {
			data.Search = append(data.Search, Keyword{data.PkgName})
		}
	}

	st := s.DB.NewStorage()
	defer st.Close()

	if err := st.Coll("subtasks").Insert(&data); err != nil {
		if storage.IsDup(err) {
			s.errorResponse(w, http.StatusBadRequest, "Task already exists")
		} else {
			s.errorResponse(w, httpStatusError(err), "%+v", err)
		}
		return
	}

	s.successResponse(w, "OK")
}

func (s *Server) apiUpdateSubtaskHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Unable to read body: %s", err)
		return
	}

	st := s.DB.NewStorage()
	defer st.Close()

	type updateSubTask struct {
		TaskID    int64
		SubTaskID int64
		Status    *string
	}

	var data updateSubTask
	if err = json.Unmarshal(msg, &data); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Invalid JSON: %s", err)
		return
	}

	changeSet := bson.M{}

	if data.Status != nil {
		changeSet["status"] = *data.Status
	}

	if len(changeSet) == 0 {
		s.errorResponse(w, http.StatusBadRequest, "More fields required")
		return
	}

	err = st.Coll("tasks").Update(bson.M{
			"taskid":    misc.ToInt64(p.Get("task")),
			"subtaskid": misc.ToInt64(p.Get("subtask")),
		}, bson.M{
			"$set": changeSet,
		})

	if err != nil {
		s.errorResponse(w, httpStatusError(err), "%+v", err)
		return
	}

	s.successResponse(w, "OK")
}

func (s *Server) apiDeleteSubtaskHandler(w *HTTPResponse, r *http.Request, p *url.Values) {
	st := s.DB.NewStorage()
	defer st.Close()

	query := bson.M{
		"taskid":    misc.ToInt64(p.Get("task")),
		"subtaskid": misc.ToInt64(p.Get("subtask")),
	}

	if err := st.Coll("subtasks").Remove(query); err != nil {
		if err != storage.ErrNotFound {
			s.errorResponse(w, httpStatusError(err), "%+v", err)
		} else {
			s.notFoundHandler(w, r, p)
		}
		return
	}

	s.successResponse(w, "OK")
}
