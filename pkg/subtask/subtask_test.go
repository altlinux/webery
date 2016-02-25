package subtask

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	//	"reflect"
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

func makeGoodSubTask() *SubTask {
	goodTask := New()

	goodTask.TaskID.Set(int64(123456))
	goodTask.SubTaskID.Set(int64(7890))
	goodTask.Owner.Set("legion")

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

func TestCheckExistence(t *testing.T) {
	data := []byte(`{
		"taskid":123456,
		"subtaskid":7890,
		"objtype":"subtask",
		"dir": "/gears/k/kernel-modules-acpi_call-un-def.git",
		"tag_name": "kernel-modules-acpi_call-un-def-0.1-alt3",
		"tag_id": "7e0b1b4a1b87aadb0dd208b9bf755116b33b52c4",
		"tag_author": "Anton V. Boyarshinov <boyarsh@altlinux>",
		"type": "repo",
		"owner": "boyarsh"
		}`)

	var task SubTask

	if err := json.Unmarshal(data, &task); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if task.CopyRepo.IsDefined() {
		t.Fatalf("Undefined field found: %+v", task)
	}

	if !task.Owner.IsDefined() {
		t.Fatalf("Defined field not found: %+v", task)
	}
}

func TestSubTaskID(t *testing.T) {
	goodTask := makeGoodSubTask()

	id, err := goodTask.GetID()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expect := bson.M{
		"taskid":    int64(123456),
		"subtaskid": int64(7890),
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
