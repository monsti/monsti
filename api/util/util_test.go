// This file is part of monsti.
// Copyright 2012-2015 Christian Neumann

// monsti/util is free software: you can redistribute it and/or modify it under
// the terms of the GNU Lesser General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.

// monsti/util is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE. See the GNU Lesser General Public License for more
// details.

// You should have received a copy of the GNU Lesser General Public License
// along with monsti/util. If not, see <http://www.gnu.org/licenses/>.
package util

import "testing"

func TestNestedMap(t *testing.T) {
	theMap := NestedMap{}
	theMap.Set("foo.bar", "hey")
	ret := theMap.Get("foo.bar")
	if ret.(string) != "hey" {
		t.Errorf("node.GetField(...) = %q, should be 'hey'", ret)
	}
}
