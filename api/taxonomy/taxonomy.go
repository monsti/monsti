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

package taxonomy

import (
	"fmt"

	"pkg.monsti.org/monsti/api/service"
)

var availableLocales = []string{"de", "en", "nl"}

// FindNodesByTerm retrieves nodes which have the given vocabulary's
// term assigned.
func FindNodesByTerm(m *service.MonstiClient, vocabulary,
	term string) ([]*service.Node, error) {
	panic("Unimplemented")
	/* TODO
	search nodes having core.Categories with the given vocabulary and term.
	return nodes
	*/
	return nil, nil
}

type Term struct {
	Name, Title, Parent string
}

// GetNodeTerms retreives all terms of the given node grouped by
// vocabulary path.
func GetNodeTerms(m *service.MonstiClient, site, nodePath string) (
	map[string][]Term, *service.CacheMods, error) {
	node, err := m.GetNode(site, nodePath)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not get node: %v", err)
	}
	ret := make(map[string][]Term)
	mods := &service.CacheMods{Deps: []service.CacheDep{{Node: nodePath}}}
	for path, terms := range node.Fields["core.Categories"].(*service.MapField).Fields {
		terms := terms.(*service.ListField).Fields
		mods.Join(&service.CacheMods{Deps: []service.CacheDep{{Node: path}}})
		vocabulary, err := m.GetNode(site, path)
		if err != nil {
			return nil, nil, fmt.Errorf("Could not get vocabulary: %v", err)
		}
		for term, attrField := range vocabulary.Fields["core.VocabularyTerms"].(*service.MapField).Fields {
			ignore := true
			for _, v := range terms {
				if v.(*service.TextField).Value().(string) == term {
					ignore = false
					break
				}
			}
			if ignore {
				continue
			}
			attributes := attrField.(*service.CombinedField).Fields
			ret[path] = append(ret[path], Term{
				Name:  term,
				Title: attributes["Title"].Value().(string),
				//				Parent: attributes["Parent"].Value().(string),
			})
		}
	}
	return ret, mods, nil
}
