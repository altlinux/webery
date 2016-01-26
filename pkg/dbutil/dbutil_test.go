package dbutil

import (
	"fmt"
	"reflect"
	"testing"

	"gopkg.in/mgo.v2/bson"

	kwd "github.com/altlinux/webery/pkg/keywords"
)

type Task struct {
	ObjType    string        `json:"objtype"`
	TimeCreate int64         `json:"timecreate"`
	Search     []kwd.Keyword `json:"search"`
	TaskID     *int64        `json:"taskid"   webery:"public,keyword"`
	Try        *int64        `json:"try"      webery:"public"`
	Iter       *int64        `json:"iter"     webery:"public"`
	Owner      *string       `json:"owner"    webery:"public"`
	State      *string       `json:"state"    webery:"public,lowercase"`
	Repo       *string       `json:"repo"     webery:"public,lowercase,keyword"`
	Aborted    *string       `json:"aborted"  webery:"public,lowercase"`
	Shared     *bool         `json:"shared"   webery:"public"`
	Swift      *bool         `json:"swift"    webery:"public"`
	TestOnly   *bool         `json:"testonly" webery:"public"`
}

func TestSerialize(t *testing.T) {
	id := int64(12345)
	try := int64(10)
	iter := int64(1)
	repo := "sisyphus"
	owner := "LegioN"
	testonly := true
	kwords := []string{"taskid", "repo"}

	x := &Task{
		TaskID:   &id,
		Repo:     &repo,
		Owner:    &owner,
		Try:      &try,
		Iter:     &iter,
		TestOnly: &testonly,
	}

	valueOf := reflect.ValueOf(x)

	b, err := UpdateBSON(valueOf)
	if err != nil {
		t.Fatalf("unable to marchal struct")
	}

	expectBSON := []bson.M{
		bson.M{
			"$pull": bson.M{
				"search": bson.M{
					"group": bson.M{
						"$in": kwords,
					},
				},
			},
		},
		bson.M{
			"$addToSet": bson.M{
				"search": bson.M{
					"$each": []kwd.Keyword{
						{
							Group: "taskid",
							Key:   fmt.Sprintf("%d", id),
						},
						{
							Group: "repo",
							Key:   repo,
						},
					},
				},
			},
			"$set": bson.M{
				"taskid":   id,
				"try":      try,
				"iter":     iter,
				"repo":     repo,
				"owner":    owner,
				"testonly": testonly,
			},
		},
	}

	if !reflect.DeepEqual(b, expectBSON) {
		t.Errorf("Wrong answer")
		fmt.Printf("expected: %+v\n", expectBSON)
		fmt.Printf("got:      %+v\n", b)
	}
}
