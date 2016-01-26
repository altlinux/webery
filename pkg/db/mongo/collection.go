package mongo

import (
	"gopkg.in/mgo.v2"

	"github.com/altlinux/webery/pkg/db"
)

type MongoCollection struct {
	*mgo.Collection
}

func (c *MongoCollection) Name() string {
	return c.Collection.FullName
}

func (c *MongoCollection) Find(query interface{}) db.Query {
	return &MongoQuery{Query: c.Collection.Find(query)}
}

func (c *MongoCollection) FindId(query interface{}) db.Query {
	return &MongoQuery{Query: c.Collection.FindId(query)}
}
