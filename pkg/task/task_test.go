package task

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"gopkg.in/mgo.v2/bson"

	"github.com/altlinux/webery/pkg/config"
	"github.com/altlinux/webery/pkg/db"
	storage "github.com/altlinux/webery/pkg/db/dbtest"
)

func showTasks(sess db.Session) {
	coll, err := sess.Coll(CollName)
	if err != nil {
		panic(err)
	}

	var result []bson.M

	err = coll.Find(bson.M{}).All(&result)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", result)
}

func initTasks(sess db.Session) {
	coll, err := sess.Coll(CollName)
	if err != nil {
		panic(err)
	}

	coll.Insert(bson.M{
		"taskid":   149239,
		"objtype":  "task",
		"try":      0,
		"iter":     1,
		"state":    "new",
		"repo":     "sisyphus",
		"owner":    "legion",
		"aborted":  "",
		"shared":   false,
		"swift":    false,
		"testonly": true,
	})
}

func TestParse(t *testing.T) {
	cfg := &config.Config{}

	dbi := storage.NewSession(cfg.Mongo)
	defer dbi.Close()

	initTasks(dbi)
	showTasks(dbi)

	taskId := int64(149239)

	task, err := GetTask(dbi, taskId)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	id, ok := task.TaskID.Get()
	if !ok {
		t.Errorf("TaskID not found: %+v", task)
	}

	if id != taskId {
		t.Errorf("Wrong task found: %+v", task)
	}
}

func TestCheckExistence(t *testing.T) {
	data := []byte(`{
		"taskid":123456,
		"objtype":"task",
		"try":0,
		"iter":1,
		"state":"new",
		"repo":"sisyphus",
		"owner":"legion",
		"testonly":true
		}`)

	var task Task

	if err := json.Unmarshal(data, &task); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	_, ok := task.Shared.Get()
	if ok {
		t.Fatalf("Undefined field found: %+v", task)
	}

	_, ok = task.TestOnly.Get()
	if !ok {
		t.Fatalf("Defined field not found: %+v", task)
	}
}

func makeGoodTask() *Task {
	goodTask := &Task{}

	goodTask.TaskID.Set(int64(123456))
	goodTask.Try.Set(int64(0))
	goodTask.Iter.Set(int64(1))
	goodTask.State.Set("new")
	goodTask.Repo.Set("sisyphus")
	goodTask.Owner.Set("legion")
	goodTask.TestOnly.Set(true)

	return goodTask
}

func TestParseJSON(t *testing.T) {
	goodTask := makeGoodTask()

	testcases := map[string]struct {
		data        []byte
		expected    *Task
		expectError bool
	}{
		"empty document": {
			data:        []byte(`{}`),
			expected:    goodTask,
			expectError: true,
		},
		"basic document": {
			data: []byte(`{
				"taskid":123456,
				"objtype":"task",
				"try":0,
				"iter":1,
				"state":"new",
				"repo":"sisyphus",
				"owner":"legion",
				"testonly":true
			}`),
			expected:    goodTask,
			expectError: false,
		},
		"uppercase strings in document": {
			data: []byte(`{
				"taskid":123456,
				"objtype":"task",
				"try":0,
				"iter":1,
				"state":"NEW",
				"repo":"SiSyPhUs",
				"owner":"LegioN",
				"testonly":true
			}`),
			expected:    goodTask,
			expectError: false,
		},
	}

	for subject, test := range testcases {
		var task Task

		if err := json.Unmarshal(test.data, &task); err != nil {
			t.Errorf("Unexpected error: %v", err)
			continue
		}

		if !reflect.DeepEqual(test.expected, &task) && !test.expectError {
			t.Errorf("unexpected diffs were produced from %s", subject)
			t.Logf("wanted: %+v", test.expected)
			t.Logf("got:    %+v", &task)
		}
	}
}
