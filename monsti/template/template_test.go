package template

import (
	"testing"
)

func TestPathJoin(t *testing.T) {
	tests := []struct {
		First, Last, Joined string
	}{
		{"", "foo", "/foo"},
		{"/", "foo", "/foo"},
		{"bar", "foo", "bar/foo"},
		{"/bar", "foo", "/bar/foo"},
		{"bar/", "foo", "bar/foo"}}
	for _, v := range tests {
		ret := pathJoin(v.First, v.Last)
		if ret != v.Joined {
			t.Errorf(`pathJoin(%q, %q) = %q, should be %q.`, v.First, v.Last, ret,
				ret)
		}
	}
}
