package db

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type dbKeySession int

const ContextSession dbKeySession = 0

var (
	ErrNotFound = mgo.ErrNotFound
	ErrCursor   = mgo.ErrCursor
)

type QueryDoc bson.M

type Iter interface {
	All(result interface{}) error
	Close() error
	Err() error
	For(result interface{}, f func() error) error
	Next(result interface{}) bool
	Timeout() bool
}

type Query interface {
	All(result interface{}) error
	One(result interface{}) error
	Count() (int, error)
	Limit(int) Query
	Skip(int) Query
	Sort(fields ...string) Query
	Iter() Iter
}

type Collection interface {
	Name() string
	Find(query interface{}) Query
	FindId(id interface{}) Query
	Insert(docs ...interface{}) error
	Remove(selector interface{}) error
	Update(selector interface{}, update interface{}) error
	Upsert(selector interface{}, update interface{}) (info *mgo.ChangeInfo, err error)
	RemoveAll(selector interface{}) (*mgo.ChangeInfo, error)
	UpdateAll(selector interface{}, update interface{}) (*mgo.ChangeInfo, error)
}

type Session interface {
	Copy() Session
	Coll(string) (Collection, error)
	Ping() error
	Close()
}

func IsDup(err error) bool {
	return mgo.IsDup(err)
}
