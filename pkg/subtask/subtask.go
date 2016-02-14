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
	ObjType    jsontype.BaseString  `json:"-,omitempty"`
	TimeCreate jsontype.Int64       `json:"-,omitempty"`
	Search     kwd.Keywords         `json:"-,omitempty"`
	TaskID     jsontype.Int64       `json:"taskid,omitempty"`
	SubTaskID  jsontype.Int64       `json:"subtaskid,omitempty"`
	Try        jsontype.Int64       `json:"try,omitempty"`
	Iter       jsontype.Int64       `json:"iter,omitempty"`
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
	id := make(db.QueryDoc)

	if tid, ok := t.TaskID.Get(); ok {
		id["taskid"] = tid
	} else {
		return id, fmt.Errorf("TaskID not specified")
	}
	if stid, ok := t.SubTaskID.Get(); ok {
		id["subtaskid"] = stid
	} else {
		return id, fmt.Errorf("SubTaskID not specified")
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

func Read(st db.Session, id int64, sid int64) (task *SubTask, err error) {
	task = &SubTask{}
	query, err := List(st, db.QueryDoc{
		"taskid":    id,
		"subtaskid": sid,
	})
	if err == nil {
		query.One(&task)
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

	if t.Search == nil {
		t.Search = kwd.NewKeywords()
	}

	if v, ok := t.Owner.Get(); ok {
		t.Search["owner"] = v
	}

	if v, ok := t.PkgName.Get(); ok {
		t.Search["pkgname"] = v
	}

	_, err = col.Upsert(id, t)
	return err
}

func Delete(st db.Session, t *SubTask) error {
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
