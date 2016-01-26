package dbtest

import (
	"gopkg.in/mgo.v2"

	"github.com/altlinux/webery/pkg/db"
)

type TestCollection struct {
	name string
	Handlers map[string]TestHandler
}

func (c *TestCollection) Name() string {
	return c.name
}

func (c *TestCollection) Find(query interface{}) db.Query {
	return &TestQuery{}
}

func (c *TestCollection) FindId(query interface{}) db.Query {
	return &TestQuery{}
}

func (c *TestCollection) Insert(docs ...interface{}) error {
	return nil
}

func (c *TestCollection) Remove(selector interface{}) error {
	return nil
}

func (c *TestCollection) Update(selector interface{}, update interface{}) error {
	return nil
}

func (c *TestCollection) RemoveAll(selector interface{}) (*mgo.ChangeInfo, error) {
	return &mgo.ChangeInfo{}, nil
}

func (c *TestCollection) UpdateAll(selector interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	return &mgo.ChangeInfo{}, nil
}

func (c *TestCollection) Upsert(selector interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	return &mgo.ChangeInfo{}, nil
}

func NewCollection(name string) *TestCollection {
	return &TestCollection{
		name:     name,
		Handlers: make(map[string]TestHandler),
	}
}