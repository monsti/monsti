// This file is part of Monsti, a web content management system.
// Copyright 2012-2014 Christian Neumann
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

package main

import (
	"fmt"
	"sort"

	"pkg.monsti.org/monsti/api/service"
	"pkg.monsti.org/monsti/api/util/template"
)

type nodeSort struct {
	Nodes  []*service.Node
	Sorter func(left, right *service.Node) bool
}

func (s *nodeSort) Len() int {
	return len(s.Nodes)
}

func (s *nodeSort) Swap(i, j int) {
	s.Nodes[i], s.Nodes[j] = s.Nodes[j], s.Nodes[i]
}

func (s *nodeSort) Less(i, j int) bool {
	return s.Sorter(s.Nodes[i], s.Nodes[j])
}

func renderNodeList(c *reqContext, context template.Context,
	h *nodeHandler) error {
	nodes, err := c.Serv.Data().GetChildren(c.Site.Name, c.Node.Path)
	if err != nil {
		return fmt.Errorf("Could not get children of node: %v", err)
	}
	order := func(left, right *service.Node) bool {
		return left.Order < right.Order
	}
	sorter := nodeSort{nodes, order}
	sort.Sort(&sorter)
	context["Children"] = nodes
	return nil
}
