package dbtest

import (
	"fmt"

	"github.com/altlinux/webery/pkg/config"
	"github.com/altlinux/webery/pkg/db"
)

type TestHandler struct {
	Process func() error
}

type TestSession struct {
	dbname string
	Handlers map[string]TestHandler
}

func (s *TestSession) Ping() error {
	return nil
}

func (s *TestSession) Close() {
	return
}

func (s *TestSession) Copy() db.Session {
	return &TestSession{}
}

func (s *TestSession) Coll(name string) (db.Collection, error) {
	coll := NewCollection(fmt.Sprintf("%s.%s", s.dbname, name))
	coll.Handlers = s.Handlers
	return coll, nil
}

func NewSession(conf config.ConfigMongo) *TestSession {
	return &TestSession{
		dbname:   conf.Database,
		Handlers: make(map[string]TestHandler),
	}
}
