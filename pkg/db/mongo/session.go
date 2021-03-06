package mongo

import (
	"regexp"
	"runtime"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"

	"github.com/altlinux/webery/pkg/config"
	"github.com/altlinux/webery/pkg/db"
)

var mongoServerDebugRe = regexp.MustCompile(`^Ping for .* is [0-9][0-9]? ms$`)

var mongoClusterDebugRe = regexp.MustCompile(`^` +
  `(SYNC Starting full topology synchronization\.\.\.` +
  `|SYNC Processing .*` +
  `|SYNC Synchronization was complete \(got data from primary\)\.` +
  `|SYNC Synchronization completed: [0-9]+ master\(s\) and [0-9]+ slave\(s\) alive\.` +
  `)$`)

type MongoSession struct {
	*mgo.Session
}

func (s MongoSession) Copy() db.Session {
	return &MongoSession{Session: s.Session.Copy()}
}

func (s MongoSession) Coll(name string) (db.Collection, error) {
	return &MongoCollection{Collection: s.Session.DB("").C(name)}, nil
}

type mongoDBLogger struct{}

func (l *mongoDBLogger) Output(calldepth int, s string) error {
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 0
	}

	if idx := strings.LastIndex(file, "/"); idx != -1 {
		file = file[idx+1:]
	}

	debug := false
	if file == "server.go" && mongoServerDebugRe.MatchString(s) {
		debug = true
	} else if file == "cluster.go" && mongoClusterDebugRe.MatchString(s) {
		debug = true
	}

	if debug {
		log.Debugf("mongodb: %s:%d: %s", file, line, s)
	} else {
		log.Infof("mongodb: %s:%d: %s", file, line, s)
	}
	return nil
}

func NewSession(conf config.Mongo) *MongoSession {
	mgo.SetLogger(&mongoDBLogger{})
	mgo.SetDebug(false)

	ss := &MongoSession{}

	for {
		var err error
		ss.Session, err = mgo.DialWithInfo(&mgo.DialInfo{
			Addrs:    conf.Hosts,
			Direct:   conf.Direct,
			FailFast: true,
			Database: conf.Database,
			Username: conf.User,
			Password: conf.Password,
			Timeout:  10 * time.Second,
		})
		if err == nil {
			break
		}

		log.Error("Unable to open MongoDB session: ", err)
		time.Sleep(1 * time.Second)
	}

	ss.SetPoolLimit(conf.PoolLimit)
	ss.SetSyncTimeout(1 * time.Minute)
	ss.SetSocketTimeout(1 * time.Minute)
	ss.SetSafe(&mgo.Safe{})

	return ss
}
