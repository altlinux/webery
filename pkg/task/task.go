package task

import (
	"fmt"

	"github.com/altlinux/webery/pkg/db"
	"github.com/altlinux/webery/pkg/jsontype"
	kwd "github.com/altlinux/webery/pkg/keywords"
)

var (
	CollName = "tasks"
)

type Task struct {
	dbname string

	ObjType    jsontype.BaseString  `json:"-,omitempty"`
	TimeCreate jsontype.Int64       `json:"-,omitempty"`
	Search     kwd.Keywords         `json:"-,omitempty"`
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

func New() *Task {
	t := &Task{
		dbname: CollName,
	}

	t.ObjType.Set("task")
	t.Search = kwd.NewKeywords()

	return t
}

func (t Task) GetDBCollection() string {
	return t.dbname
}

func (t Task) GetID() (db.QueryDoc, error) {
	id := make(db.QueryDoc)

	if tid, ok := t.TaskID.Get(); ok {
		id["taskid"] = tid
	} else {
		return id, fmt.Errorf("TaskID not specified")
	}
	return id, nil
}

func List(st db.Session, query db.QueryDoc) (res db.Query, err error) {
	col, err := st.Coll(CollName)
	if err == nil {
		res = col.Find(query)
	}
	return
}

func Read(st db.Session, id int64) (task *Task, err error) {
	task = New()
	query, err := List(st, db.QueryDoc{"taskid": id})
	if err == nil {
		query.One(&task)
	}
	return
}

func Write(st db.Session, t *Task) error {
	col, err := st.Coll(CollName)
	if err != nil {
		return err
	}

	id, err := t.GetID()
	if err != nil {
		return err
	}

	if t.Search == nil {
		t.Search = kwd.NewKeywords()
	}

	if t.TaskID.IsDefined() {
		t.Search["taskid"] = t.TaskID.String()
	}

	if t.Repo.IsDefined() {
		t.Search["repo"] = t.Repo.String()
	}

	_, err = col.Upsert(id, t)
	return err
}

func Delete(st db.Session, t *Task) error {
	col, err := st.Coll(CollName)
	if err != nil {
		return err
	}

	id, err := t.GetID()
	if err != nil {
		return err
	}

	return col.Remove(id)
}
