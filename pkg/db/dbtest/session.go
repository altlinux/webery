package dbtest

import (
	"crypto/rand"
	"fmt"
	"os"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/dbtest"

	"github.com/altlinux/webery/pkg/config"
	"github.com/altlinux/webery/pkg/db"
	"github.com/altlinux/webery/pkg/db/mongo"
)

type TestSession struct {
	*mgo.Session

	dbname string
	dbdir  string
	Server dbtest.DBServer
}

func (s *TestSession) Close() {
	s.Session.Close()
	s.Server.Stop()
	_ = os.RemoveAll(s.dbdir)
	return
}

func (s *TestSession) Copy() db.Session {
	new_s := s
	new_s.Session = s.Session.Copy()
	return new_s
}

func (s *TestSession) Coll(name string) (db.Collection, error) {
	return &mongo.MongoCollection{Collection: s.Session.DB("").C(name)}, nil
}

func NewSession(conf config.Mongo) *TestSession {
	s := &TestSession{}
	s.dbname = "dbtest"

	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	s.dbdir = fmt.Sprintf("/tmp/%s-%x", s.dbname, b)
	if err := os.MkdirAll(s.dbdir, 0777); err != nil {
		panic(err)
	}

	s.Server.SetPath(s.dbdir)
	s.Session = s.Server.Session()

	return s
}
