package main

import (
	"datenkarussell.de/monsti/rpc/client"
	"testing"
)

func TestCheckPermission(t *testing.T) {
	tests := []struct {
		Action      string
		Auth, Grant bool
	}{
		{"", false, true},
		{"", true, true},
		{"edit", false, false},
		{"edit", true, true},
		{"add", false, false},
		{"add", true, true},
		{"unknown_action", true, false},
		{"unknown_action", false, false}}
	for _, v := range tests {
		var user *client.User
		if v.Auth {
			user = &client.User{}
		}
		ret := checkPermission(v.Action, &client.Session{User: user})
		if ret != v.Grant {
			t.Errorf("checkPermission(%v, %v) = %v, expected %v", v.Action,
				user, ret, v.Grant)
		}
	}
}
