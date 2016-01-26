package dbtest

import (
	"github.com/altlinux/webery/pkg/db"
)

type TestQuery struct {
	limit int
	skip int
	query interface{}
}

func (c *TestQuery) Count() (int, error) {
	return 1000, nil
}

func (c *TestQuery) Limit(n int) db.Query {
	return &TestQuery{
		query: c.query,
		skip: c.skip,
		limit: n,
	}
}

func (c *TestQuery) Skip(n int) db.Query {
	return &TestQuery{
		query: c.query,
		limit: c.limit,
		skip: n,
	}
}

func (c *TestQuery) Sort(fields ...string) db.Query {
	return &TestQuery{
		query: c.query,
		limit: c.limit,
		skip: c.skip,
	}
}

func (c *TestQuery) All(result interface{}) error {
	return nil
}

func (c *TestQuery) One(result interface{}) error {
	return nil
}

func (c *TestQuery) Iter() db.Iter {
	return &TestIter{}
}
