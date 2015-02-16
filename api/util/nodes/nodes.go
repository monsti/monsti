// This file is part of Monsti, a web content management system.
// Copyright 2012-2015 Christian Neumann
//
// Monsti is free software: you can redistribute it and/or modify it under the
// terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.
//
// Monsti is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
// A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
// details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Monsti.  If not, see <http://www.gnu.org/licenses/>.

/*
Package nodes implements utility functions to work with node slices.
*/

package nodes

import (
	"pkg.monsti.org/monsti/api/service"
)

// Sorter implements `sort.Interface` to allow node sorting.
type Sorter struct {
	// The node slice to sort
	Nodes []*service.Node
	// The compare function.
	LessFunc func(left, right *service.Node) bool
}

func (s *Sorter) Len() int {
	return len(s.Nodes)
}

func (s *Sorter) Swap(i, j int) {
	s.Nodes[i], s.Nodes[j] = s.Nodes[j], s.Nodes[i]
}

func (s *Sorter) Less(i, j int) bool {
	return s.LessFunc(s.Nodes[i], s.Nodes[j])
}
