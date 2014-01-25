// Copyright 2012-2013 Christian Neumann
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
// details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package service

import (
	"bytes"
	"reflect"
	"testing"
)

func TestGetConfig(t *testing.T) {
	tests := []struct {
		Body string
		Out  interface{}
		Ret  interface{}
	}{
		{`{"fookey":"foovalue"}`, "", "foovalue"},
		{`{"fookey": null}`, "", ""},
	}
	for _, test := range tests {
		err := getConfig([]byte(test.Body), &test.Out)
		if err != nil {
			t.Error("getConfig returned error: %v", err)
		}
		if !reflect.DeepEqual(test.Out, test.Ret) {
			t.Error("getConfig(%q, out); out is %q, should be %q",
				test.Body, test.Out, test.Ret)
		}
	}
}

type FooFields struct {
	FooField1 string
	FooField2 int
}

type BarFields struct {
	BarField1 string
	BarField2 int
}

type FooBarNode struct {
	FooFields
	BarFields
}

func TestNodeToData(t *testing.T) {
	var node FooBarNode
	node.FooFields = FooFields{"FooField1Val", 13}
	node.BarFields = BarFields{"BarField1Val", 4}
	data, err := nodeToData(node, []string{"foo"})
	if err != nil {
		t.Fatalf("nodeToData returns error: %v", err)
	}
	expected := []byte(`{"FooField1":"FooField1Val","FooField2":13}`)
	if len(data) != 1 {
		t.Fatalf("nodeToData should return a slice of length 1, got length %d",
			len(data))
	}
	if !bytes.Equal(data[0], expected) {
		t.Fatalf("nodeToData should return %q, got %q", expected, data[0])
	}
}

func TestDataToNode(t *testing.T) {
	var node, expected FooBarNode
	expected.FooFields = FooFields{"FooField1Val", 13}
	data := []byte(`{"FooField1":"FooField1Val","FooField2":13}`)
	err := dataToNode([][]byte{data}, &node, []string{"foo"})
	if err != nil {
		t.Fatalf("dataToNode returns error: %v", err)
	}
	if !reflect.DeepEqual(node, expected) {
		t.Fatalf("dataToNode should fill as %v, got %v", expected, node)
	}
}
