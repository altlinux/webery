package api

import (
	"github.com/altlinux/webery/pkg/db"
)

type Query struct {
	CollName string
	Pattern  db.QueryDoc
	Sort     []string
	Iterator func(db.Iter) interface{}
	One      func(db.Query) (interface{}, error)
}
