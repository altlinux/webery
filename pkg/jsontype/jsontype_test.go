package jsontype

import (
	"encoding/json"
	"reflect"
	"testing"

	"gopkg.in/mgo.v2/bson"

	kwd "github.com/altlinux/webery/pkg/keywords"
)

type Task struct {
	Foo Bool              `json:",omitempty"`
	Bar Int64             `json:",omitempty"`
	Baz BaseString        `json:",omitempty"`
	Bax LowerString       `json:",omitempty"`
	Far []kwd.Keyword     `json:",omitempty" bson:",omitempty"`
	Xxx *int              `json:",omitempty"`
	Zzz map[string]string `json:",omitempty" bson:",omitempty"`
}

func TestAssign(t *testing.T) {
	task := &Task{}

	task.Foo.Set(false)
	if v, ok := task.Foo.Get(); ok {
		if v != false {
			t.Errorf("Field 'Foo' has an unexpected value: %+v", task)
		}
	} else {
		t.Errorf("Field 'Foo' is not specified: %+v", task)
	}

	task.Foo.Set(true)
	if v, ok := task.Foo.Get(); ok {
		if v != true {
			t.Errorf("Field 'Foo' has an unexpected value: %+v", task)
		}
	} else {
		t.Errorf("Field 'Foo' is not specified: %+v", task)
	}

	task.Bar.Set(int64(1))
	if v, ok := task.Bar.Get(); ok {
		if v != int64(1) {
			t.Errorf("Field 'Bar' has an unexpected value: %+v", task)
		}
	} else {
		t.Errorf("Field 'Bar' is not specified: %+v", task)
	}

	task.Baz.Set("xxx")
	if v, ok := task.Baz.Get(); ok {
		if v != "xxx" {
			t.Errorf("Field 'Baz' has an unexpected value: %+v", task)
		}
	} else {
		t.Errorf("Field 'Baz' is not specified: %+v", task)
	}

	task.Bax.Set("XxX")
	if v, ok := task.Bax.Get(); ok {
		if v != "xxx" {
			t.Errorf("Field 'Bax' has an unexpected value: %+v", task)
		}
	} else {
		t.Errorf("Field 'Bax' is not specified: %+v", task)
	}
}

func TestBoolToString(t *testing.T) {
	task := &Task{}

	if task.Foo.String() != "<nil>" {
		t.Errorf("Field 'Foo' has an unexpected value: %s", task.Foo.String())
	}

	task.Foo.Set(false)
	if task.Foo.String() != "false" {
		t.Errorf("Field 'Foo' has an unexpected value: %s", task.Foo.String())
	}

	task.Foo.Set(true)
	if task.Foo.String() != "true" {
		t.Errorf("Field 'Foo' has an unexpected value: %s", task.Foo.String())
	}
}

func TestInt64ToString(t *testing.T) {
	task := &Task{}

	if task.Bar.String() != "<nil>" {
		t.Errorf("Field 'Bar' has an unexpected value: %s", task.Bar.String())
	}

	task.Bar.Set(int64(1))
	if task.Bar.String() != "1" {
		t.Errorf("Field 'Bar' has an unexpected value: %s", task.Bar.String())
	}
}

func TestBaseStringToString(t *testing.T) {
	task := &Task{}

	if task.Baz.String() != "<nil>" {
		t.Errorf("Field 'Baz' has an unexpected value: %s", task.Baz.String())
	}

	task.Baz.Set("xxx")
	if task.Baz.String() != "xxx" {
		t.Errorf("Field 'Baz' has an unexpected value: %s", task.Baz.String())
	}

	task.Bax.Set("XxX")
	if task.Bax.String() != "xxx" {
		t.Errorf("Field 'Bax' has an unexpected value: %s", task.Bax.String())
	}
}

func TestBSON(t *testing.T) {
	testcases := map[string]struct {
		data     *Task
		expected bson.M
	}{
		"empty document": {
			data:     &Task{},
			expected: bson.M{},
		},
		"document with only one bool=true defined": {
			data: &Task{
				Foo: *NewBool(true),
			},
			expected: bson.M{
				"foo": true,
			},
		},
		"document with only one bool=false defined": {
			data: &Task{
				Foo: *NewBool(false),
			},
			expected: bson.M{
				"foo": false,
			},
		},
		"document with only one int64 defined": {
			data: &Task{
				Bar: *NewInt64(int64(123)),
			},
			expected: bson.M{
				"bar": int64(123),
			},
		},
		"document with only one string defined": {
			data: &Task{
				Baz: *NewBaseString("zzz"),
			},
			expected: bson.M{
				"baz": "zzz",
			},
		},
		"document with two fields defined": {
			data: &Task{
				Foo: *NewBool(true),
				Bar: *NewInt64(int64(123)),
			},
			expected: bson.M{
				"foo": true,
				"bar": int64(123),
			},
		},
		"document with list of keywords": {
			data: &Task{
				Foo: *NewBool(true),
				Far: []kwd.Keyword{
					kwd.Keyword{
						Key:   "AAA",
						Group: "repo",
					},
				},
			},
			expected: bson.M{
				"foo": true,
				"far": []bson.M{
					bson.M{"key": "AAA", "group": "repo"},
				},
			},
		},
	}

	for title, test := range testcases {
		expectOut, err := bson.Marshal(&test.expected)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultOut, err := bson.Marshal(&test.data)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultTask := &Task{}
		if err := bson.Unmarshal(resultOut, resultTask); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expectTask := &Task{}
		if err := bson.Unmarshal(expectOut, expectTask); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(resultTask, expectTask) {
			t.Errorf("Unexpected difference in %s", title)

			t.Logf("expected: %+v\n", expectTask)
			t.Logf("got     : %+v\n", resultTask)

			b0, err := json.Marshal(expectTask)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			b1, err := json.Marshal(resultTask)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			t.Logf("expected: %+v\n", string(b0))
			t.Logf("got     : %+v\n", string(b1))
		}
	}
}

func TestJSON(t *testing.T) {
	testcases := map[string]struct {
		data     *Task
		expected interface{}
	}{
		"empty document": {
			data:     &Task{},
			expected: struct{}{},
		},
		"document with only one bool=true defined": {
			data: &Task{
				Foo: *NewBool(true),
			},
			expected: struct {
				Foo bool
			}{
				Foo: true,
			},
		},
		"document with only one bool=false defined": {
			data: &Task{
				Foo: *NewBool(false),
			},
			expected: struct {
				Foo bool
			}{
				Foo: false,
			},
		},
		"document with only one string defined": {
			data: &Task{
				Baz: *NewBaseString("zzz"),
			},
			expected: struct {
				Baz string
			}{
				Baz: "zzz",
			},
		},
		"document with two fields defined": {
			data: &Task{
				Foo: *NewBool(true),
				Bar: *NewInt64(int64(123)),
			},
			expected: struct {
				Foo bool
				Bar int64
			}{
				Foo: true,
				Bar: 123,
			},
		},
		"document with list of keywords": {
			data: &Task{
				Foo: *NewBool(true),
				Far: []kwd.Keyword{
					kwd.Keyword{
						Key:   "AAA",
						Group: "repo",
					},
				},
			},
			expected: struct {
				Foo bool
				Far []struct {
					Key   string
					Group string
				}
			}{
				Foo: true,
				Far: []struct {
					Key   string
					Group string
				}{
					{
						Key:   "AAA",
						Group: "repo",
					},
				},
			},
		},
	}

	for title, test := range testcases {
		expectOut, err := json.Marshal(&test.expected)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultOut, err := json.Marshal(&test.data)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		var resultTask Task
		if err := json.Unmarshal(resultOut, &resultTask); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		var expectTask Task
		if err := json.Unmarshal(expectOut, &expectTask); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(&resultTask, &expectTask) {
			t.Errorf("Unexpected difference in %s", title)

			t.Logf("expected: %+v\n", &expectTask)
			t.Logf("got     : %+v\n", &resultTask)

			t.Logf("expected (json): %+v\n", string(expectOut))
			t.Logf("got      (json): %+v\n", string(resultOut))
		}
	}
}
