// This file is part of Monsti, a web content management system.
// Copyright 2012-2013 Christian Neumann
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

/* Packaga nodeutil provides utility functions for monsti nodes */
package nodeutil

import (
	"encoding/json"
	"fmt"
	"net/url"

	"pkg.monsti.org/service"
)

// SimpleJSONRequest receives the JSON data of the given node into out.
func SimpleJSONRequest(path string, s *service.Session, req *service.Request,
	out interface{}) error {
	url, err := url.Parse(path)
	if err != nil {
		return fmt.Errorf("nodeutil: Could not parse node path: %v", err)
	}
	node, err := s.Data().GetNode(req.Site, url.Path)
	if err != nil {
		return fmt.Errorf("nodeutil: Could not find node: %v", err)
	}
	nodeServ, err := s.Info().FindNodeService(node.Type)
	if err != nil {
		return fmt.Errorf(
			"nodeutil: Could not find node service for %q at %q: %v",
			node.Type, url.Path, err)
	}
	defer func() {
		if err := nodeServ.Close(); err != nil {
			panic(fmt.Sprintf("nodeutil: Could not close node service connection:",
				err))
		}
	}()
	subReq := service.Request{
		Site:    req.Site,
		Method:  service.GetRequest,
		Node:    *node,
		Query:   url.Query(),
		Session: req.Session,
		Action:  service.ViewAction,
	}
	res, err := nodeServ.Request(&subReq)
	if err != nil {
		return fmt.Errorf("nodeutil: Could not request node: %v", err)
	}
	if len(res.Redirect) > 0 {
		return fmt.Errorf("nodeutil: Got redirect to %v", res.Redirect)
	}
	err = json.Unmarshal(res.Body, out)
	if err != nil {
		return fmt.Errorf("nodeutil: Could not unmarshall body: %v", err)
	}
	return nil
}
