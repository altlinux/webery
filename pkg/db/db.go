package db

import (
	"errors"
	"gopkg.in/mgo.v2"
)

var (
	ErrNotFound = errors.New("not found")
	ErrCursor   = errors.New("invalid cursor")
)

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
