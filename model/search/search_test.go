package search

import (
	"testing"
)

func TestNewKeywords(t *testing.T) {
	k := NewKeywords()

	k.Append("grp", "111")
	k.Append("grp", "222")
	k.Append("grp", "333")

	if k.Length() != 3 {
		t.Fatalf("wrong length = %d, expected 3", k.Length())
	}

}
