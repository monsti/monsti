package main

import (
	"io/ioutil"
	"launchpad.net/goyaml"
	"path/filepath"
)

// getFooter retrieves the footer.
//
// root is the path to the data directory
//
// Returns an empty string if there is no footer.
func getFooter(root string) string {
	path := filepath.Join(root, "footer.html")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

// getBelowHeader retrieves the below header content for the given node.
//
// path is the node's path.
// root is the path to the data directory.
//
// Returns an empty string if there is no below header content.
func getBelowHeader(path, root string) string {
	file := filepath.Join(root, path, "below_header.html")
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return ""
	}
	return string(content)
}

// getSidebar retrieves the sidebar content for the given node.
//
// path is the node's path.
// root is the path to the data directory.
//
// It traverses up to the root until it finds a node with defined sidebar
// content.
//
// Returns an empty string if there is no sidebar content.
func getSidebar(path, root string) string {
	for {
		file := filepath.Join(root, path, "sidebar.html")
		content, err := ioutil.ReadFile(file)
		if err != nil {
			if path == filepath.Dir(path) {
				break
			}
			path = filepath.Dir(path)
			continue
		}
		return string(content)
	}
	return ""
}

// navLink represents a link in the navigation.
type navLink struct {
	Name, Target string
	Active       bool
}

// getNav returns the navigation for the given node.
// 
// node is the path of the node for which to get the navigation.
// active is the currently active node.
// root is the path of the data directory.
//
// The keys of the returned map are the link titles, the values are
// the link targets.
func getNav(node, active, root string) []navLink {
	path := filepath.Join(root, node, "navigation.yaml")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	var navLinks []navLink
	goyaml.Unmarshal(content, &navLinks)
        for i, link := range navLinks {
		if link.Target == active {
			navLinks[i].Active = true
			break
		}
	}
	return navLinks
}

