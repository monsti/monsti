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
 Monsti is a simple and resource efficient CMS.

 This package implements the document node type.
*/
package main

import (
	"fmt"
	"html/template"
	"log"
	"net/url"

	"sort"
	"strconv"
	"pkg.monsti.org/monsti/api/service"
	"pkg.monsti.org/monsti/api/util/i18n"
	"pkg.monsti.org/monsti/api/util/nodes"
	mtemplate "pkg.monsti.org/monsti/api/util/template"
)

func getBlogPosts(req *service.Request, blogPath string, s *service.Session,
	limit int) ([]*service.Node, error) {
	var posts []*service.Node
	years, err := s.Monsti().GetChildren(req.Site, blogPath)
	if err != nil {
		return nil, fmt.Errorf("Could not fetch year children: %v", err)
	}
	for _, year := range years {
		months, err := s.Monsti().GetChildren(req.Site, year.Path)
		if err != nil {
			return nil, fmt.Errorf("Could not fetch month children: %v", err)
		}
		for _, month := range months {
			monthPosts, err := s.Monsti().GetChildren(req.Site, month.Path)
			if err != nil {
				return nil, fmt.Errorf("Could not fetch month children: %v", err)
			}
			posts = append(posts, monthPosts...)
		}
	}
	order := func(left, right *service.Node) bool {
		return left.PublishTime.Before(right.PublishTime)
	}
	sort.Sort(sort.Reverse(&nodes.Sorter{posts, order}))
	return posts, nil
}

func getBlogContext(reqId uint, embed *service.EmbedNode,
	s *service.Session, settings *settings, renderer *mtemplate.Renderer) (
	map[string]interface{}, *service.CacheMods, error) {
	req, err := s.Monsti().GetRequest(reqId)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not get request: %v", err)
	}
	query := req.Query
	blogPath := req.NodePath
	if embed != nil {
		embedUrl, err := url.Parse(embed.URI)
		if err != nil {
			return nil, nil, fmt.Errorf("Could not parse embed URI")
		}
		query = embedUrl.Query()
		blogPath = embedUrl.Path
	}
	limit := -1
	if limitParam, err := strconv.Atoi(query.Get("limit")); err == nil {
		limit = limitParam
		if limit < 1 {
			limit = 1
		}
	}
	context := mtemplate.Context{}
	context["Embedded"] = embed
	context["Posts"], err = getBlogPosts(req, blogPath, s, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not retrieve blog posts: %v", err)
	}
	rendered, err := renderer.Render("core/blogpost-list", context,
		req.Session.Locale, settings.Monsti.GetSiteTemplatesPath(req.Site))
	if err != nil {
		return nil, nil, fmt.Errorf("Could not render template: %v", err)
	}
	mods := &service.CacheMods{
		Deps: []service.CacheDep{{Node: req.NodePath, Descend: 1}},
	}
	return map[string]interface{}{"BlogPosts": template.HTML(string(rendered))}, mods, nil
}

func initBlog(settings *settings, session *service.Session,
	sessions *service.SessionPool, logger *log.Logger,
	renderer *mtemplate.Renderer) error {
	G := func(in string) string { return in }

	nodeType := service.NodeType{
		Id:        "core.Blog",
		AddableTo: []string{"."},
		Name:      i18n.GenLanguageMap(G("Blog"), availableLocales),
		Fields: []*service.FieldConfig{
			{Id: "core.Title"},
		},
	}
	if err := session.Monsti().RegisterNodeType(&nodeType); err != nil {
		return fmt.Errorf("Could not register blog node type: %v", err)
	}

	nodeType = service.NodeType{
		Id:        "core.BlogPost",
		AddableTo: []string{"core.Blog"},
		Name:      i18n.GenLanguageMap(G("Blog Post"), availableLocales),
		Fields: []*service.FieldConfig{
			{Id: "core.Title"},
			{Id: "core.Body"},
		},
		Hide:       true,
		PathPrefix: "$year/$month",
	}
	if err := session.Monsti().RegisterNodeType(&nodeType); err != nil {
		return fmt.Errorf("Could not register blog post node type: %v", err)
	}

	// Add a signal handler
	handler := service.NewNodeContextHandler(sessions,
		func(req uint, session *service.Session, nodeType string,
			embedNode *service.EmbedNode) (
			map[string]interface{}, *service.CacheMods, error) {
			switch nodeType {
			case "core.Blog":
				ctx, mods, err := getBlogContext(req, embedNode, session, settings, renderer)
				if err != nil {
					return nil, nil, fmt.Errorf("Could not get blog context: %v", err)
				}
				return ctx, mods, nil
			default:
				return nil, nil, nil
			}
		})
	if err := session.Monsti().AddSignalHandler(handler); err != nil {
		logger.Fatalf("Could not add signal handler: %v", err)
	}
	return nil
}
