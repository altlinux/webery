package task

import (
	"fmt"

	"github.com/altlinux/webery/pkg/db"
	"github.com/altlinux/webery/pkg/jsontype"
	kwd "github.com/altlinux/webery/pkg/keywords"
	evtype "github.com/altlinux/webery/pkg/taskevents"
)

var (
	CollName = "tasks"
)

type TaskEvent struct {
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

func NewTaskEvent() *TaskEvent {
	return &TaskEvent{}
}

type Task struct {
	Search     kwd.Keywords         `json:"-"`
	Events     *evtype.EventList    `json:"events,omitempty"`
	ObjType    jsontype.BaseString  `json:"objtype,omitempty"`
	TimeCreate jsontype.Int64       `json:"timecreate,omitempty"`
	TaskID     jsontype.Int64       `json:"taskid,omitempty"`
	Owner      jsontype.LowerString `json:"owner,omitempty"`
	State      jsontype.LowerString `json:"state,omitempty"`
	Repo       jsontype.LowerString `json:"repo,omitempty"`
	Aborted    jsontype.LowerString `json:"aborted,omitempty"`
	Shared     jsontype.Bool        `json:"shared,omitempty"`
	Swift      jsontype.Bool        `json:"swift,omitempty"`
	TestOnly   jsontype.Bool        `json:"testonly,omitempty"`
}

func New() *Task {
	t := &Task{}

	t.ObjType.Set("task")
	t.Search = kwd.NewKeywords()
	t.Events = evtype.NewEventList()

	return t
}

func (t Task) GetID() (db.QueryDoc, error) {
	tid, ok := t.TaskID.Get()
	if !ok {
		return nil, fmt.Errorf("TaskID not specified")
	}
	return MakeID(tid), nil
}

func MakeID(id int64) db.QueryDoc {
	res := make(db.QueryDoc)
	res["taskid"] = id
	return res
}

func List(st db.Session, query db.QueryDoc) (res db.Query, err error) {
	col, err := st.Coll(CollName)
	if err == nil {
		res = col.Find(query)
	}
	return
}

func Read(st db.Session, id db.QueryDoc) (task *Task, err error) {
	task = New()
	query, err := List(st, id)
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

	t.ObjType.Set("task")

	if t.Search == nil {
		t.Search = kwd.NewKeywords()
	}

	if t.Events == nil {
		t.Events = evtype.NewEventList()
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

func Delete(st db.Session, query db.QueryDoc) error {
	col, err := st.Coll(CollName)
	if err != nil {
		return err
	}

	_, err = col.RemoveAll(query)
	return err
}
