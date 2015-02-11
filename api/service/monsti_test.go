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
	"reflect"
	"strings"
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
		{``, "", ""},
	}
	for _, test := range tests {
		err := getConfig([]byte(test.Body), &test.Out)
		if err != nil {
			t.Errorf("getConfig returned error: %v", err)
		}
		if !reflect.DeepEqual(test.Out, test.Ret) {
			t.Errorf("getConfig(%q, out); out is %q, should be %q",
				test.Body, test.Out, test.Ret)
		}
	}
}

func TestDataToNode(t *testing.T) {
	nodeType := NodeType{
		Id:   "foo.Bar",
		Name: map[string]string{"en": "A Bar"},
		Fields: []*NodeField{{
			Id:   "foo.FooField",
			Name: map[string]string{"en": "A FooField"},
			Type: "Text",
		}},
		Embed: nil}
	data := []byte(`
{ "Type": "foo.Bar",
  "Fields": {
    "foo": {
      "FooField": "Foo Value",
      "BarField": "Bar Value"
    }
  },
  "LocalFields": [
		{
			"Id": "foo.BarField",
			"Name": null,
			"Required": false,
			"Type": "Text"
		}
	]
}`)
	getNodeType := func(id string) (*NodeType, error) { return &nodeType, nil }
	node, err := dataToNode(data, getNodeType, nil, "")
	if err != nil {
		t.Fatalf("dataToNode returns error: %v", err)
	}
	ret := node.Fields["foo.FooField"].Value().(string)
	if ret != "Foo Value" {
		t.Errorf(`Field foo.FooField = %q, should be "Foo Value"`, ret)
	}
	ret = node.Fields["foo.BarField"].Value().(string)
	if ret != "Bar Value" {
		t.Errorf(`Field foo.BarField = %q, should be "Bar Value"`, ret)
	}
}

func TestNodeToData(t *testing.T) {
	node := Node{
		Path: "/foo",
		Type: &NodeType{
			Id: "foo.Bar",
			Fields: []*NodeField{{
				Id:   "foo.FooField",
				Type: "Text",
			}},
		},
		LocalFields: []*NodeField{{Id: "foo.BarField", Type: "Text"}},
	}
	node.InitFields(nil, "")
	f := func(in interface{}) error {
		*(in.(*TextField)) = "FooValue"
		return nil
	}
	node.Fields["foo.FooField"].Load(f)
	f = func(in interface{}) error {
		*(in.(*TextField)) = "BarValue"
		return nil
	}
	node.Fields["foo.BarField"].Load(f)

	expected := `{
		  "Order": 0,
		  "Hide": false,
		  "TemplateOverwrites": null,
		  "Embed": null,
		  "LocalFields": [
				{
					"Id": "foo.BarField",
					"Name": null,
					"Required": false,
          "Hidden": false,
					"Type": "Text"
				}
			],
		  "Public": false,
		  "PublishTime": "0001-01-01T00:00:00Z",
      "Changed":"0001-01-01T00:00:00Z",
		  "Type": "foo.Bar",
		  "Fields": {
		    "foo": {
		      "BarField": "BarValue",
		      "FooField": "FooValue"
		    }
		  }
}`

	oldPath := node.Path
	ret, err := nodeToData(&node, true)
	if oldPath != node.Path {
		t.Errorf("nodeToData altered node")
	}
	if err != nil {
		t.Errorf("Got error: %v", err)
	}
	trim := func(in string) string {
		out := strings.Replace(in, " ", "", -1)
		out = strings.Replace(out, "\n", "", -1)
		out = strings.Replace(out, "\t", "", -1)
		return out
	}
	if trim(string(ret)) != trim(expected) {
		t.Errorf("nodeToData(%v, true) is\n`%v`\n, should be\n`%v`", node,
			trim(string(ret)), trim(expected))
	}
}
