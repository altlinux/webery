package subtask

import (
	"fmt"

	"github.com/altlinux/webery/pkg/db"
	"github.com/altlinux/webery/pkg/jsontype"
	kwd "github.com/altlinux/webery/pkg/keywords"
)

var (
	CollName = "subtasks"
)

type SubTask struct {
	Search     kwd.Keywords         `json:"-"`
	ObjType    jsontype.BaseString  `json:"objtype,omitempty"`
	TimeCreate jsontype.Int64       `json:"timecreate,omitempty"`
	TaskID     jsontype.Int64       `json:"taskid,omitempty"`
	SubTaskID  jsontype.Int64       `json:"subtaskid,omitempty"`
	Owner      jsontype.LowerString `json:"owner,omitempty"`
	Type       jsontype.LowerString `json:"type,omitempty"`
	Status     jsontype.LowerString `json:"status,omitempty"`
	CopyRepo   jsontype.LowerString `json:"copyrepo,omitempty"`
	Package    jsontype.BaseString  `json:"package,omitempty"`
	Srpm       jsontype.BaseString  `json:"srpm,omitempty"`
	Project    jsontype.BaseString  `json:"project,omitempty"`
	Dir        jsontype.BaseString  `json:"dir,omitempty"`
	PkgName    jsontype.BaseString  `json:"pkgname,omitempty"`
	TagName    jsontype.BaseString  `json:"tagname,omitempty"`
	TagAuthor  jsontype.LowerString `json:"tagauthor,omitempty"`
	TagID      jsontype.LowerString `json:"tagid,omitempty"`
}

func New() *SubTask {
	t := &SubTask{}

	t.ObjType.Set("subtask")
	t.Search = kwd.NewKeywords()

	return t
}

func (t SubTask) GetID() (db.QueryDoc, error) {
	tid, ok := t.TaskID.Get()
	if !ok {
		return nil, fmt.Errorf("TaskID not specified")
	}

	stid, ok := t.SubTaskID.Get()
	if !ok {
		return nil, fmt.Errorf("SubTaskID not specified")
	}

	return MakeID(tid, stid), nil
}

func (t SubTask) IsCancelled() bool {
	status, ok := t.Status.Get()
	if !ok || status != "cancelled" {
		return false
	}
	return true
}

func MakeID(id int64, sid int64) db.QueryDoc {
	res := make(db.QueryDoc)
	res["taskid"] = id
	res["subtaskid"] = sid
	return res
}

func List(st db.Session, query db.QueryDoc) (res db.Query, err error) {
	col, err := st.Coll(CollName)
	if err == nil {
		res = col.Find(query)
	}
	return
}

func Read(st db.Session, id db.QueryDoc) (subtask *SubTask, err error) {
	subtask = New()
	query, err := List(st, id)
	if err == nil {
		query.One(&subtask)
	}
	return
}

func Write(st db.Session, t *SubTask) error {
	col, err := st.Coll(CollName)
	if err != nil {
		return err
	}

	id, err := t.GetID()
	if err != nil {
		return err
	}

	t.ObjType.Set("subtask")

	if t.Search == nil {
		t.Search = kwd.NewKeywords()
	}

	if t.Owner.IsDefined() {
		t.Search["owner"] = t.Owner.String()
	}

	if t.PkgName.IsDefined() {
		t.Search["pkgname"] = t.PkgName.String()
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
