package keywords

import (
	//"fmt"
	"testing"

	//"github.com/altlinux/webery/pkg/model"
)

func TestNewKeywords(t *testing.T) {
	k := NewKeywords()

	data := map[string]string{
		"grp0": "111",
		"grp1": "222",
		"grp2": "333",
	}

	for n, v := range data {
		k.Append(n, v)
	}

	if k.Length() != 3 {
		t.Fatalf("wrong length = %d, expected 3", k.Length())
	}

	for n, v := range data {
		kwd, ok := k.Group(n)
		if !ok {
			t.Fatalf("'%s' not found", n)
		}
		if kwd.Group != n {
			t.Fatalf("wrong group = '%s', expected '%s'", kwd.Group, n)
		}
		if kwd.Key != v {
			t.Fatalf("wrong key = '%s', expected '%s'", kwd.Key, v)
		}
	}
}

func TestReplaceGroup(t *testing.T) {
	grpName := "grp"

	k := NewKeywords()
	k.Append(grpName, "111")
	k.Append(grpName, "222")

	kwd, ok := k.Group(grpName)
	if !ok {
		t.Fatalf("'%s' not found", grpName)
	}

	if kwd.Key != "222" {
		t.Fatalf("wrong key = '%s', expected '222'", kwd.Key)
	}
}

func TestReplaceGroupByEmptyString(t *testing.T) {
	grpName := "grp"

	k := NewKeywords()
	k.Append(grpName, "111")
	k.Append(grpName, "")

	kwd, ok := k.Group(grpName)
	if !ok {
		t.Fatalf("'%s' not found", grpName)
	}

	if kwd.Key != "111" {
		t.Fatalf("wrong key = '%s', expected '111'", kwd.Key)
	}
}

func TestAddEmptyString(t *testing.T) {
	grpName := "grp"

	k := NewKeywords()
	k.Append(grpName, "")

	_, ok := k.Group(grpName)
	if ok {
		t.Fatalf("'%s' found, but should not", grpName)
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
		k.Append(n, v)
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
		k.Append(n, v)
	}

	grps := k.Groups()

	for _, n := range grps {
		_, ok := data[n]
		if !ok {
			t.Fatalf("'%s' not found in data", n)
		}
		if n == "grp3" {
			t.Fatalf("'%s' empty group found", n)
		}
	}
}
