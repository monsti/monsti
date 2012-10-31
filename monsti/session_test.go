package main

import (
	"datenkarussell.de/monsti/rpc/client"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCheckPermission(t *testing.T) {
	tests := []struct {
		Action      string
		Auth, Grant bool
	}{
		{"", false, true},
		{"", true, true},
		{"login", false, true},
		{"login", true, true},
		{"logout", false, false},
		{"logout", true, true},
		{"edit", false, false},
		{"edit", true, true},
		{"add", false, false},
		{"add", true, true},
		{"remove", false, false},
		{"remove", true, true},
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

func TestGetUser(t *testing.T) {
	root, err := ioutil.TempDir("", "_monsti_get_user")
	if err != nil {
		t.Fatalf("Could not create temp dir: %s", err)
	}
	db := []byte(`- login: foo
  name: Mr. Foo
  password: the pass
  email: foo@example.com
- login: bar
  name: Mrs. Bar
  email: bar@example.com
  password: other pass
`)
	if err = ioutil.WriteFile(filepath.Join(root, "users.yaml"),
		db, 0600); err != nil {
		t.Fatalf("Could not write navigation: %s", err)
	}
	tests := []struct {
		Login string
		User  *client.User
	}{
		{Login: "unknown", User: nil},
		{Login: "foo", User: &client.User{Login: "foo", Password: "the pass",
			Name: "Mr. Foo", Email: "foo@example.com"}},
		{Login: "bar", User: &client.User{Login: "bar", Password: "other pass",
			Name: "Mrs. Bar", Email: "bar@example.com"}}}
	for _, v := range tests {
		user := getUser(v.Login, root)
		if !reflect.DeepEqual(user, v.User) {
			t.Errorf("getUser(%q, _) = %v, should be %v", v.Login,
				user, v.User)
		}
	}
}
