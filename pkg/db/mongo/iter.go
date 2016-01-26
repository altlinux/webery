package mongo

import (
	"gopkg.in/mgo.v2"
)

type MongoIter struct {
	*mgo.Iter
}
