/*
 * Copyright (C) 2015 Alexey Gladkov <gladkov.alexey@gmail.com>
 *
 * This file is covered by the GNU General Public License,
 * which should be included with webery as the file COPYING.
 */

package storage

import (
	"expvar"
	//	"runtime"
	//	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"

	"github.com/altlinux/webery/config"
)

var (
	ErrNotFound = mgo.ErrNotFound
)

type mongoDBLogger struct{}

func (l *mongoDBLogger) Output(calldepth int, s string) error {
	/*
		_, file, line, ok := runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}

		if idx := strings.LastIndex(file, "/"); idx != -1 {
			file = file[idx+1:]
		}

		log.Infof("mongodb: %s:%d: %s", file, line, s)
	*/
	return nil
}

type MongoService struct {
	globalSession *mgo.Session
}

func (ss *MongoService) NewStorage() *MongoStorage {
	return &MongoStorage{
		session: ss.globalSession.Copy(),
	}
}

type MongoStorage struct {
	session *mgo.Session
}

func (st *MongoStorage) Ping() error {
	return st.session.Ping()
}

func (st *MongoStorage) Refresh() {
	st.session.Refresh()
}

func (st *MongoStorage) Close() {
	st.session.Close()
}

func (st *MongoStorage) Coll(name string) *mgo.Collection {
	return st.session.DB("").C(name)
}

func NewMongoService(conf config.ConfigMongo) *MongoService {
	mgo.SetLogger(&mongoDBLogger{})
	//mgo.SetDebug(true)

	ss := &MongoService{}

	for {
		var err error
		ss.globalSession, err = mgo.DialWithInfo(&mgo.DialInfo{
			Addrs:    conf.Hosts,
			Direct:   conf.Direct,
			FailFast: true,
			Database: conf.Database,
			Username: conf.User,
			Password: conf.Password,
			Timeout:  10 * time.Second,
		})
		if err == nil {
			break
		}

		log.Error("Unable to open MongoDB session: ", err)
		time.Sleep(1 * time.Second)
	}

	ss.globalSession.SetPoolLimit(conf.PoolLimit)
	ss.globalSession.SetSyncTimeout(1 * time.Minute)
	ss.globalSession.SetSocketTimeout(1 * time.Minute)
	ss.globalSession.SetSafe(&mgo.Safe{})

	return ss
}

func (st *MongoStorage) Initialize() {
	index := mgo.Index{
		Key:        []string{"taskid"},
		Unique:     true,
		DropDups:   true,
		Background: false,
		Sparse:     false,
	}
	if err := st.Coll("tasks").EnsureIndex(index); err != nil {
		log.Fatalf("mongodb: %#v", err)
	}

	index = mgo.Index{
		Key:        []string{"search.key"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     true,
	}
	if err := st.Coll("tasks").EnsureIndex(index); err != nil {
		log.Fatalf("mongodb: %#v", err)
	}

	index = mgo.Index{
		Key:        []string{"taskid", "subtaskid"},
		Unique:     true,
		DropDups:   true,
		Background: false,
		Sparse:     false,
	}
	if err := st.Coll("subtasks").EnsureIndex(index); err != nil {
		log.Fatalf("mongodb: %#v", err)
	}

	index = mgo.Index{
		Key:        []string{"search.key"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     true,
	}
	if err := st.Coll("subtasks").EnsureIndex(index); err != nil {
		log.Fatalf("mongodb: %#v", err)
	}

	index = mgo.Index{
		Key:        []string{"repo", "name"},
		Unique:     true,
		DropDups:   true,
		Background: false,
		Sparse:     false,
	}
	if err := st.Coll("acl_packages").EnsureIndex(index); err != nil {
		log.Fatalf("mongodb: %#v", err)
	}

	index = mgo.Index{
		Key:        []string{"repo", "name"},
		Unique:     true,
		DropDups:   true,
		Background: false,
		Sparse:     false,
	}
	if err := st.Coll("acl_groups").EnsureIndex(index); err != nil {
		log.Fatalf("mongodb: %#v", err)
	}

	index = mgo.Index{
		Key:        []string{"user"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	}
	if err := st.Coll("git_packages").EnsureIndex(index); err != nil {
		log.Fatalf("mongodb: %#v", err)
	}
}

func MongoExpvar() {
	mgo.SetStats(true)
	expvar.Publish("mongodb", expvar.Func(func() interface{} {
		return mgo.GetStats()
	}))
}

func IsDup(err error) bool {
	return mgo.IsDup(err)
}
