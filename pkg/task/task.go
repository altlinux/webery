package task

import (
	"reflect"

	"gopkg.in/mgo.v2/bson"

	"github.com/altlinux/webery/pkg/db"
	"github.com/altlinux/webery/pkg/dbutil"
	kwd "github.com/altlinux/webery/pkg/keywords"
)

var (
	CollName = "tasks"
)

type Task struct {
	ObjType    string        `json:"objtype"`
	TimeCreate int64         `json:"timecreate"`
	Search     []kwd.Keyword `json:"search"`
	TaskID     *int64        `json:"taskid"   webery:"public,keyword"`
	Try        *int64        `json:"try"      webery:"public"`
	Iter       *int64        `json:"iter"     webery:"public"`
	Owner      *string       `json:"owner"    webery:"public"`
	State      *string       `json:"state"    webery:"public,lowercase"`
	Repo       *string       `json:"repo"     webery:"public,lowercase,keyword"`
	Aborted    *string       `json:"aborted"  webery:"public,lowercase"`
	Shared     *bool         `json:"shared"   webery:"public"`
	Swift      *bool         `json:"swift"    webery:"public"`
	TestOnly   *bool         `json:"testonly" webery:"public"`
}

func (t *Task) getUpdateBSON() ([]bson.M, error) {
	valueOf := reflect.ValueOf(t)
	return dbutil.UpdateBSON(valueOf)
}

func GetTask(st db.Session, ID int64) (*Task, error) {
	col, err := st.Coll(CollName)
	if err != nil {
		return nil, err
	}

	t := &Task{}
	col.Find(bson.M{"taskid": ID}).One(t)

	return t, nil
}
