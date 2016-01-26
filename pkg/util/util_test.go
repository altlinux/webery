/*
 * Copyright (C) 2015 Alexey Gladkov <gladkov.alexey@gmail.com>
 *
 * This file is covered by the GNU General Public License,
 * which should be included with webery as the file COPYING.
 */

package util

import "testing"

func TestToInt32(t *testing.T) {
	chks := map[string]int32{
		"-12":                 -12,
		"123":                 123,
		"255":                 255,
		"32768":               32768,
		"65535":               65535,
		"2147483647":          2147483647,
		"-2147483647":         -2147483647,
		"4294967295":          0,
		"9223372036854775807": 0,
		"foo": 0,
		"":    0,
	}

	for k, v := range chks {
		i := ToInt32(k)
		if i != v {
			t.Fatalf("wrong answer = %d, expected %s", i, k)
		}
	}
}

func TestToInt64(t *testing.T) {
	chks := map[string]int64{
		"-12":                  -12,
		"123":                  123,
		"255":                  255,
		"32768":                32768,
		"65535":                65535,
		"2147483647":           2147483647,
		"4294967295":           4294967295,
		"9223372036854775807":  9223372036854775807,
		"-9223372036854775807": -9223372036854775807,
		"18446744073709551615": 0,
		"foo": 0,
		"":    0,
	}

	for k, v := range chks {
		i := ToInt64(k)
		if i != v {
			t.Fatalf("wrong answer = %d, expected %s", i, k)
		}
	}
}

func TestInSliceString(t *testing.T) {
	chks := []string{"aaa", "bbb", "ccc"}

	for _, s := range chks {
		if !InSliceString(s, chks) {
			t.Fatalf("can not find '%s'", s)
		}
	}

	if InSliceString("xxx", chks) {
		t.Fatalf("find 'xxx'")
	}
}

func TestInSliceStringNilList(t *testing.T) {
	if InSliceString("xxx", nil) {
		t.Fatalf("find 'xxx'")
	}
}
