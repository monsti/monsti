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

package service

/*

import (
	"github.com/chrneumann/mimemail"
	"github.com/monsti/rpc/types"
	"io"
	"log"
	"net/rpc"
	"net/url"
	"os"
	"strings"
)

// GetFileData retrieves the content of the uploaded file with the given key. It
// has to be called before any calls to GetFormData.
func (s *Client) GetFileData(key string) ([]byte, error) {
	var reply []byte
	err := s.Call("NodeRPC.GetFileData", &key, &reply)
	return reply, err
}

// GetFormData retrieves form data of the request, i.e. query string values and
// possibly form data of POST and PUT requests.
func (s *Client) GetFormData() url.Values {
	var reply url.Values
	err := s.Call("NodeRPC.GetFormData", 0, &reply)
	if err != nil {
		s.Logger.Fatal("Master: RPC GetFormData error:", err)
	}
	return reply
}
*/
