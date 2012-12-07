package main

import (
	"testing"
)

func TestInStringSlice(t *testing.T) {
	tests := []struct {
		value    string
		elements []string
		in       bool
	}{
		{"foo", []string{"foo", "bar"}, true},
		{"bar", []string{"foo", "bar"}, true},
		{"", []string{"", "bar"}, true},
		{"bla", []string{"foo", "bar"}, false},
		{"foo", []string{"bar"}, false},
		{"foo", []string{}, false},
		{"", []string{}, false},
		{"", []string{"foo", "bar"}, false}}
	for _, test := range tests {
		if ret := inStringSlice(test.value, test.elements); ret != test.in {
			t.Errorf("inStringSlice(%q, %v) = %v, should be %v",
				test.value, test.elements, ret, !ret)
		}
	}
}
