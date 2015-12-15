/*
 * Copyright (C) 2015 Alexey Gladkov <gladkov.alexey@gmail.com>
 *
 * This file is covered by the GNU General Public License,
 * which should be included with webery as the file COPYING.
 */

package subtask

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/altlinux/webery/config"
	"github.com/altlinux/webery/misc"
	"github.com/altlinux/webery/model/search"
	"github.com/altlinux/webery/storage"
)

const collection = "subtasks"

type SubTask struct {
	ObjType    string           `json:"objtype"    field:"private"`
	TimeCreate int64            `json:"timecreate" field:"private"`
	Search     []search.Keyword `json:"search"     field:"private"`
	TaskID     *int64           `json:"taskid"     field:"public" type:"int"`
	SubTaskID  *int64           `json:"subtaskid"  field:"public" type:"int"`
	Owner      *string          `json:"owner"      field:"public" type:"string" textcase:"lower"`
	Type       *string          `json:"type"       field:"public" type:"string" textcase:"lower"`
	Status     *string          `json:"status"     field:"public" type:"string" textcase:"lower"`
	Package    *string          `json:"package"    field:"public" type:"string"`
	Srpm       *string          `json:"srpm"       field:"public" type:"string"`
	CopyRepo   *string          `json:"copyrepo"   field:"public" type:"string" textcase:"lower"`
	Project    *string          `json:"project"    field:"public" type:"string"`
	Dir        *string          `json:"dir"        field:"public" type:"string"`
	TagAuthor  *string          `json:"tagauthor"  field:"public" type:"string" textcase:"lower"`
	TagID      *string          `json:"tagid"      field:"public" type:"string" textcase:"lower"`
	TagName    *string          `json:"tagname"    field:"public" type:"string"`
	PkgName    *string          `json:"pkgname"    field:"public" type:"string"`
}

func Valid(t SubTask) error {
	cfg := config.GetConfig()

	if t.TaskID != nil && *t.TaskID <= 0 {
		return fmt.Errorf("Bad TaskID")
	}

	if t.SubTaskID != nil && *t.SubTaskID <= 0 {
		return fmt.Errorf("Bad SubTaskID")
	}

	if t.Type != nil && !misc.InSliceString(*t.Type, cfg.Builder.SubTaskTypes) {
		return fmt.Errorf("Wrong value for 'type' field")
	}

	if t.Status != nil && !misc.InSliceString(*t.Status, cfg.Builder.SubTaskStates) {
		return fmt.Errorf("Wrong value for 'status' field")
	}

	if t.CopyRepo != nil && !misc.InSliceString(*t.CopyRepo, cfg.Builder.Repos) {
		return fmt.Errorf("Unknown repo: %s", *t.CopyRepo)
	}

	return nil
}

func IsSubTaskCancelled(data SubTask) bool {
	if data.Status != nil && *data.Status == "cancelled" {
		return true
	}
	return false
}

func ListByTaskID(st *storage.MongoStorage, taskID int64) *mgo.Query {
	return st.Coll(collection).Find(bson.M{
		"taskid": taskID,
	})
}

func GetSubTask(st *storage.MongoStorage, taskID int64, subtaskID int64) (t *SubTask, err error) {
	t = &SubTask{}
	err = st.Coll(collection).
		Find(bson.M{
		"taskid":    taskID,
		"subtaskid": subtaskID,
	}).
		One(t)
	return
}

func Create(st *storage.MongoStorage, t SubTask) error {
	typeOfTask := reflect.TypeOf(t)
	valueOfTask := reflect.Indirect(reflect.ValueOf(&t))

	for i := 0; i < typeOfTask.NumField(); i++ {
		field := typeOfTask.Field(i)

		if field.Tag.Get("field") == "private" {
			continue
		}

		value := valueOfTask.Field(i)

		if value.Kind() == reflect.Ptr && value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
	}

	kwds := search.NewKeywords()
	kwds.Append("owner", *t.Owner)
	kwds.Append("pkgname", *t.PkgName)

	if t.Dir != nil && len(*t.Dir) > 0 {
		project := strings.TrimSuffix(filepath.Base(*t.Dir), ".git")
		kwds.Append("project", project)
	}

	t.ObjType = "subtask"
	t.TimeCreate = time.Now().Unix()
	t.Search = kwds.Keywords()

	return st.Coll(collection).Insert(t)
}

func RemoveByID(st *storage.MongoStorage, taskID int64, subtaskID int64) error {
	return st.Coll(collection).Remove(bson.M{
		"taskid":    taskID,
		"subtaskid": subtaskID,
	})
}

func RemoveByTaskID(st *storage.MongoStorage, taskID int64) error {
	return st.Coll(collection).Remove(bson.M{
		"taskid": taskID,
	})
}

func UpdateSubTask(st *storage.MongoStorage, taskID int64, subtaskID int64, data SubTask) error {
	kwds := search.NewKeywords()
	kwds.Append("taskid", fmt.Sprintf("%d", taskID))

	strFields := make(map[string]*string)
	changeSet := bson.M{}

	typeOfTask := reflect.TypeOf(data)
	valueOfTask := reflect.ValueOf(data)

	for i := 0; i < typeOfTask.NumField(); i++ {
		field := typeOfTask.Field(i)

		if field.Tag.Get("field") == "private" {
			continue
		}

		value := valueOfTask.Field(i)

		if value.IsNil() {
			continue
		}

		name := field.Tag.Get("json")

		if field.Tag.Get("type") == "string" {
			s := value.Interface().(*string)

			if field.Tag.Get("textcase") == "lower" {
				*s = strings.ToLower(*s)
			}

			strFields[name] = s
			changeSet[name] = s

			continue
		}

		if name == "taskid" {
			tmp := fmt.Sprintf("%d", value.Int())
			strFields[name] = &tmp
		}

		changeSet[name] = value.Interface()
	}

	if data.Dir != nil {
		project := strings.TrimSuffix(filepath.Base(*data.Dir), ".git")
		strFields["dir"] = &project
	}

	for k, v := range strFields {
		switch k {
		case "owner", "pkgname", "project":
			kwds.Append(k, *v)
		}
	}

	search := bson.M{
		"taskid":    taskID,
		"subtaskid": subtaskID,
	}
	change := bson.M{}

	if len(changeSet) > 0 {
		change["$set"] = changeSet
	}

	if kwds.Length() > 0 {
		change["$addToSet"] = bson.M{
			"search": bson.M{
				"$each": kwds.Keywords(),
			},
		}
		removeKeywords := bson.M{
			"$pull": bson.M{
				"search": bson.M{
					"group": bson.M{
						"$in": kwds.Groups(),
					},
				},
			},
		}

		if err := st.Coll(collection).Update(search, removeKeywords); err != nil {
			return err
		}
	}

	if len(change) == 0 {
		return nil
	}

	return st.Coll(collection).Update(search, change)
}
