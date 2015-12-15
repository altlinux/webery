/*
 * Copyright (C) 2015 Alexey Gladkov <gladkov.alexey@gmail.com>
 *
 * This file is covered by the GNU General Public License,
 * which should be included with webery as the file COPYING.
 */

package task

import (
	"fmt"
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

const collection = "tasks"

type Task struct {
	ObjType    string           `json:"objtype"    field:"private"`
	TimeCreate int64            `json:"timecreate" field:"private"`
	Search     []search.Keyword `json:"search"     field:"private"`
	TaskID     *int64           `json:"taskid"     field:"public" type:"int"`
	Try        *int64           `json:"try"        field:"public" type:"int"`
	Iter       *int64           `json:"iter"       field:"public" type:"int"`
	State      *string          `json:"state"      field:"public" type:"string" textcase:"lower"`
	Repo       *string          `json:"repo"       field:"public" type:"string" textcase:"lower"`
	Owner      *string          `json:"owner"      field:"public" type:"string" textcase:"lower"`
	Aborted    *string          `json:"aborted"    field:"public" type:"string" textcase:"lower"`
	Shared     *bool            `json:"shared"     field:"public" type:"bool"`
	Swift      *bool            `json:"swift"      field:"public" type:"bool"`
	TestOnly   *bool            `json:"testonly"   field:"public" type:"bool"`
}

func Valid(t Task) error {
	cfg := config.GetConfig()

	if t.TaskID != nil && *t.TaskID <= 0 {
		return fmt.Errorf("Bad TaskID")
	}

	if t.Repo != nil && !misc.InSliceString(*t.Repo, cfg.Builder.Repos) {
		return fmt.Errorf("Unknown repo: %s", t.Repo)
	}

	if t.State != nil && !misc.InSliceString(*t.State, cfg.Builder.TaskStates) {
		return fmt.Errorf("Wrong value for 'status' field")
	}

	return nil
}

func List(st *storage.MongoStorage, filter bson.M) *mgo.Query {
	return st.Coll(collection).Find(filter)
}

func GetTask(st *storage.MongoStorage, ID int64) (t *Task, err error) {
	t = &Task{}
	err = st.Coll(collection).
		Find(bson.M{"taskid": ID}).
		One(t)
	return
}

func Create(st *storage.MongoStorage, t Task) error {
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
	kwds.Append("taskid", fmt.Sprintf("%d", *t.TaskID))
	kwds.Append("repo", *t.Repo)

	t.ObjType = "task"
	t.TimeCreate = time.Now().Unix()
	t.Search = kwds.Keywords()

	return st.Coll(collection).Insert(t)
}

func RemoveByID(st *storage.MongoStorage, taskID int64) error {
	return st.Coll(collection).Remove(bson.M{
		"taskid": taskID,
	})
}

func UpdateTask(st *storage.MongoStorage, taskID int64, data Task) error {
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

	for k, v := range strFields {
		switch k {
		case "repo", "taskid":
			kwds.Append(k, *v)
		}
	}

	search := bson.M{
		"taskid": taskID,
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
