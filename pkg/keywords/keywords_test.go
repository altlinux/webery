package keywords

import (
	"reflect"
	"testing"

	"gopkg.in/mgo.v2/bson"
)

func TestNewKeywords(t *testing.T) {
	k := NewKeywords()

	data := map[string]string{
		"grp0": "111",
		"grp1": "222",
		"grp2": "333",
	}

	for n, v := range data {
		k[n] = v
	}

	if len(k) != 3 {
		t.Fatalf("wrong length = %d, expected 3", len(k))
	}

	for n, v := range data {
		kwd, ok := k[n]
		if !ok {
			t.Fatalf("'%s' not found", n)
		}
		if kwd != v {
			t.Fatalf("wrong key = '%s', expected '%s'", kwd, v)
		}
	}
}

func TestReplaceGroup(t *testing.T) {
	grpName := "grp"

	k := NewKeywords()
	k[grpName] = "111"
	k[grpName] = "222"

	kwd, ok := k[grpName]
	if !ok {
		t.Fatalf("'%s' not found", grpName)
	}

	if kwd != "222" {
		t.Fatalf("wrong key = '%s', expected '222'", kwd)
	}
}

func TestReturnKeywords(t *testing.T) {
	k := NewKeywords()

	data := map[string]string{
		"grp0": "111",
		"grp1": "222",
		"grp2": "333",
		"grp3": "",
	}

	for n, v := range data {
		k[n] = v
	}

	kwds := k.Keywords()

	for _, kwd := range kwds {
		v, ok := data[kwd.Group]
		if !ok {
			t.Fatalf("'%s' not found in data", kwd.Group)
		}
		if kwd.Key != v {
			t.Fatalf("wrong key = '%s', expected '%s'", v, kwd.Key)
		}
	}
}

func TestReturnGroups(t *testing.T) {
	k := NewKeywords()

	data := map[string]string{
		"grp0": "111",
		"grp1": "222",
		"grp2": "333",
		"grp3": "",
	}

	for n, v := range data {
		k[n] = v
	}

	grps := k.Groups()

	for _, n := range grps {
		_, ok := data[n]
		if !ok {
			t.Fatalf("'%s' not found in data", n)
		}
	}
}

type fakeDoc struct {
	Search Keywords
}

func TestSetBSON(t *testing.T) {
	d0 := &fakeDoc{
		Search: NewKeywords(),
	}
	d0.Search["foo"] = "xxx"
	d0.Search["bar"] = "123"

	d1 := &fakeDoc{
		Search: NewKeywords(),
	}

	doc := bson.M{
		"search": []bson.M{
			{"group": "foo", "key": "xxx"},
			{"group": "bar", "key": "123"},
		},
	}

	out, err := bson.Marshal(doc)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	err = bson.Unmarshal(out, d1)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if !reflect.DeepEqual(d0, d1) {
		t.Errorf("Unexpected difference")

		t.Logf("expected: %+v\n", d0)
		t.Logf("got     : %+v\n", d1)
	}
}
