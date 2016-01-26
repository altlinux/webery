package mongo

import (
	"gopkg.in/mgo.v2"

	"github.com/altlinux/webery/pkg/db"
)

type MongoQuery struct {
	*mgo.Query
}

func (c *MongoQuery) Limit(n int) db.Query {
	return &MongoQuery{Query: c.Query.Limit(n)}
}

func (c *MongoQuery) Skip(n int) db.Query {
	return &MongoQuery{Query: c.Query.Skip(n)}
}

func (c *MongoQuery) Sort(fields ...string) db.Query {
	return &MongoQuery{Query: c.Query.Sort(fields...)}
}

func (c *MongoQuery) Iter() db.Iter {
	return &MongoIter{Iter: c.Query.Iter()}
}
