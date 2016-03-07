package task

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"testing"

	"gopkg.in/mgo.v2/bson"

	"github.com/altlinux/webery/pkg/config"
	"github.com/altlinux/webery/pkg/db"
	storage "github.com/altlinux/webery/pkg/db/dbtest"
)

var (
	dbi db.Session
)

func TestMain(m *testing.M) {
	flag.Parse()

	cfg := &config.Config{}

	dbi = storage.NewSession(cfg.Mongo)
	defer dbi.Close()

	os.Exit(m.Run())
}

func makeGoodTask() *Task {
	goodTask := New()

	goodTask.TaskID.Set(int64(123456))
	goodTask.Try.Set(int64(0))
	goodTask.Iter.Set(int64(1))
	goodTask.State.Set("new")
	goodTask.Repo.Set("sisyphus")
	goodTask.Owner.Set("legion")
	goodTask.TestOnly.Set(true)

	return goodTask
}

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

func cleanTasks(sess db.Session) {
	coll, err := sess.Coll(CollName)
	if err != nil {
		panic(err)
	}
	coll.RemoveAll(bson.M{})
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

	if task.Shared.IsDefined() {
		t.Fatalf("Undefined field found: %+v", task)
	}

	if !task.TestOnly.IsDefined() {
		t.Fatalf("Defined field not found: %+v", task)
	}
}

func TestTaskID(t *testing.T) {
	goodTask := makeGoodTask()

	id, err := goodTask.GetID()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expect := bson.M{
		"taskid": int64(123456),
	}

	expectOut, err := bson.Marshal(&expect)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resultOut, err := bson.Marshal(id)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !bytes.Equal(expectOut, resultOut) {
		t.Errorf("Unexpected diff")
		t.Logf("wanted: %+v", expect)
		t.Logf("got:    %+v", id)
	}
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
		task := New()

		if err := json.Unmarshal(test.data, task); err != nil {
			t.Errorf("Unexpected error: %v", err)
			continue
		}

		if !reflect.DeepEqual(test.expected, task) && !test.expectError {
			t.Errorf("unexpected diffs were produced from %s", subject)
			t.Logf("wanted: %+v", test.expected)
			t.Logf("got:    %+v", task)
		}
	}
}

func TestRead(t *testing.T) {
	initTasks(dbi)
	//	showTasks(dbi)

	taskId := int64(149239)

	task, err := Read(dbi, MakeID(taskId))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	id, ok := task.TaskID.Get()
	if !ok {
		t.Errorf("TaskID not found: %+v", task)
	}

	if id != taskId {
		t.Errorf("Wrong task found: %+v", task)
	}

	cleanTasks(dbi)
}

func TestWrite(t *testing.T) {
	badTask := New()
	if err := Write(dbi, badTask); err == nil {
		t.Fatalf("Expected error")
	}

	goodTask := makeGoodTask()

	taskId, _ := goodTask.TaskID.Get()

	if err := Write(dbi, goodTask); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	task, err := Read(dbi, MakeID(taskId))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	id, ok := task.TaskID.Get()
	if !ok {
		t.Errorf("TaskID not found: %+v", task)
	}

	if id != taskId {
		t.Errorf("Wrong task found: %+v", task)
	}

	cleanTasks(dbi)
}

func TestUpdate(t *testing.T) {
	goodTask := makeGoodTask()

	taskId, _ := goodTask.TaskID.Get()

	if err := Write(dbi, goodTask); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	goodTask.Search = nil
	goodTask.Owner.Set("ldv")

	if err := Write(dbi, goodTask); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	task, err := Read(dbi, MakeID(taskId))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	owner, ok := task.Owner.Get()
	if !ok {
		t.Errorf("Owner not found: %+v", task)
	}

	if owner != "ldv" {
		t.Errorf("Wrong task found: %+v", task)
	}

	cleanTasks(dbi)
}

func TestDelete(t *testing.T) {
	goodTask := makeGoodTask()

	gootTaskID, _ := goodTask.TaskID.Get()

	if err := Write(dbi, goodTask); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	task, err := Read(dbi, MakeID(gootTaskID))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	id, ok := task.TaskID.Get()
	if !ok {
		t.Errorf("TaskID not found: %+v", task)
	}

	if id != gootTaskID {
		t.Errorf("Wrong task found: %+v", task)
	}

	if err := Delete(dbi, MakeID(id)); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	_, err = Read(dbi, MakeID(id))
	if err != nil {
		if err != db.ErrNotFound {
			t.Fatalf("Unexpected error: %v", err)
		}
	}
}
