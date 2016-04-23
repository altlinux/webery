package taskevents

import (
	"encoding/json"
	"reflect"
	"testing"

	//"gopkg.in/mgo.v2/bson"
)

func TestNewEventList(t *testing.T) {
	k := NewEventList()

	k.Append("111")
	k.Append("222")
	k.Append("333")
	k.Append("111")

	if k.Length() != 3 {
		t.Fatalf("wrong length = %d, expected 3", k.Length())
	}
}


func TestDefined(t *testing.T) {
	k := NewEventList()

	if k.IsDefined() {
		t.Fatalf("misidentified")
	}

	k.Append("111")

	if !k.IsDefined() {
		t.Fatalf("wrongly not defined")
	}
}

func TestString(t *testing.T) {
	k := NewEventList()

	if k.String() != "<nil>" {
		t.Fatalf("wrong string value: %s", k.String())
	}

	k.Append("111")
	k.Append("222")

	if k.String() != "[\"111\" \"222\"]" {
		t.Fatalf("wrong string value: %s", k.String())
	}
}

func TestReadonly(t *testing.T) {
	k := NewEventList()

	k.Append("111")
	k.Append("222")
	k.Readonly(true)
	k.Append("333")

	if k.Length() != 2 {
		t.Fatalf("wrong length = %d, expected 2", k.Length())
	}
}

func TestSet(t *testing.T) {
	k := NewEventList()

	k.Append("111")
	k.Append("222")
	k.Append("333")

	k.Set([]string{"000"})

	if k.Length() != 1 {
		t.Fatalf("wrong length = %d, expected 1", k.Length())
	}

	k.Readonly(true)
	k.Set([]string{"555"})

	v, ok := k.Get()
	if !ok {
		t.Fatalf("unexpected miss")
	}

	if v[0] != "000" {
		t.Fatalf("wrong value: got %q, expected '000'", v[0])
	}
}

func TestGet(t *testing.T) {
	k := NewEventList()

	k.Append("111")
	k.Append("222")
	k.Append("333")

	v, ok := k.Get()
	if !ok {
		t.Fatalf("unexpected miss")
	}

	for i, e := range []string{"111","222","333"} {
		if v[i] != e {
			t.Fatalf("wrong value: got %q, expected %q", v[i], e)
		}
	}
}

func TestUnmarshalJSON(t *testing.T) {
	data := []byte(`["111","222","333"]`)

	n := NewEventList()
	n.Readonly(true)

	if err := json.Unmarshal(data, n); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if n.Length() != 0 {
		t.Fatalf("wrong length = %d, expected 0", n.Length())
	}

	n.Readonly(false)

	if err := json.Unmarshal(data, n); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	k := NewEventList()
	k.Append("111")
	k.Append("222")
	k.Append("333")

	if !reflect.DeepEqual(n, k) {
		t.Errorf("unexpected diffs")
		t.Logf("wanted: %s", k.String())
		t.Logf("got:    %s", n.String())
	}
}

func TestMarshalJSON(t *testing.T) {
	k := NewEventList()

	exp := []byte(`null`)
	res, err := json.Marshal(k)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	k.Append("111")
	k.Append("222")
	k.Append("333")

	exp = []byte(`["111","222","333"]`)
	res, err = json.Marshal(k)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(exp, res) {
		t.Errorf("unexpected diffs")
	}
}
