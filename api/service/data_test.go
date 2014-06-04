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
