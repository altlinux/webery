package task

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/altlinux/webery/pkg/db"
	"github.com/altlinux/webery/pkg/jsontype"
	kwd "github.com/altlinux/webery/pkg/keywords"
)

var (
	CollName = "tasks"
)

type Task struct {
	ObjType    jsontype.BaseString  `json:"-,omitempty"`
	TimeCreate jsontype.Int64       `json:"-,omitempty"`
	Search     []kwd.Keyword        `json:"-,omitempty"`
	TaskID     jsontype.Int64       `json:"taskid,omitempty"`
	Try        jsontype.Int64       `json:"try,omitempty"`
	Iter       jsontype.Int64       `json:"iter,omitempty"`
	Owner      jsontype.LowerString `json:"owner,omitempty"`
	State      jsontype.LowerString `json:"state,omitempty"`
	Repo       jsontype.LowerString `json:"repo,omitempty"`
	Aborted    jsontype.LowerString `json:"aborted,omitempty"`
	Shared     jsontype.Bool        `json:"shared,omitempty"`
	Swift      jsontype.Bool        `json:"swift,omitempty"`
	TestOnly   jsontype.Bool        `json:"testonly,omitempty"`
}

func GetTask(st db.Session, ID int64) (*Task, error) {
	col, err := st.Coll(CollName)
	if err != nil {
		return nil, err
	}

	t := &Task{}
	if err := col.Find(bson.M{"taskid": ID}).One(t); err != nil {
		return nil, err
	}

	return t, nil
}

func Create(st db.Session, t Task) error {
	kwds := kwd.NewKeywords()

	if v, ok := t.TaskID.Get(); ok {
		kwds.Append("taskid", fmt.Sprintf("%d", v))
	}

	if v, ok := t.Repo.Get(); ok {
		kwds.Append("repo", v)
	}

	t.Search = kwds.Keywords()
	t.ObjType.Set("task")
	t.TimeCreate.Set(time.Now().Unix())

	col, err := st.Coll(CollName)
	if err != nil {
		return err
	}
	return col.Insert(t)
}

func RemoveByID(st db.Session, ID int64) error {
	col, err := st.Coll(CollName)
	if err != nil {
		return err
	}
	return col.Remove(bson.M{"taskid": ID})
}
