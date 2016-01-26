package subtask

import (
	"reflect"

	"gopkg.in/mgo.v2/bson"

	kwd "github.com/altlinux/webery/pkg/keywords"
	"github.com/altlinux/webery/pkg/dbutil"
)

type SubTask struct {
	ObjType    string        `json:"objtype"`
	TimeCreate int64         `json:"timecreate"`
	Search     []kwd.Keyword `json:"search"`
	TaskID     *int64        `json:"taskid"    webery:"public"`
	SubTaskID  *int64        `json:"subtaskid" webery:"public"`
	Owner      *string       `json:"owner"     webery:"public,lowercase,keyword"`
	Type       *string       `json:"type"      webery:"public,lowercase"`
	Status     *string       `json:"status"    webery:"public,lowercase"`
	Package    *string       `json:"package"   webery:"public"`
	Srpm       *string       `json:"srpm"      webery:"public"`
	CopyRepo   *string       `json:"copyrepo"  webery:"public,lowercase"`
	Project    *string       `json:"project"   webery:"public,keyword"`
	Dir        *string       `json:"dir"       webery:"public"`
	TagAuthor  *string       `json:"tagauthor" webery:"public,lowercase"`
	TagID      *string       `json:"tagid"     webery:"public,lowercase"`
	TagName    *string       `json:"tagname"   webery:"public"`
	PkgName    *string       `json:"pkgname"   webery:"public,keyword"`
}

func (t *SubTask) getUpdateBSON() ([]bson.M, error) {
	valueOf := reflect.ValueOf(t)
	return dbutil.UpdateBSON(valueOf)
}
