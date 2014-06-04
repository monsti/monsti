// Tool to upgrade the node files
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"pkg.monsti.org/monsti/api/service"
)

type OldNode struct {
	Type  string
	Order int
	Hide  bool
	Title string
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Println("Please specify path to nodes")
		os.Exit(1)
	}

	walker := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		log.Println("Visiting node ", path)
		nodeJSON, err := ioutil.ReadFile(filepath.Join(path, "node.json"))
		if err != nil {
			return fmt.Errorf("Could not read node.json: %v", err)
		}
		var oldNode OldNode
		var node service.Node
		err = json.Unmarshal(nodeJSON, &oldNode)
		if err != nil {
			return fmt.Errorf("Could not unmarshol node.json: %v", err)
		}
		if oldNode.Type == "Image" {
			oldNode.Type = "File"
			img, _ := ioutil.ReadFile(filepath.Join(path, "image.data"))
			ioutil.WriteFile(filepath.Join(path, "__file_core.File"), img, 0644)
		}
		body, err := ioutil.ReadFile(filepath.Join(path, "body.html"))
		if err == nil {
			body = bytes.Replace(body, []byte("/?raw=1"), []byte(""), -1)
			node.SetField("core.Body", string(body))
			log.Println("Wrote core.body")
		}
		node.Type = "core." + oldNode.Type
		node.Order = oldNode.Order
		node.Hide = oldNode.Hide
		node.SetField("core.Title", oldNode.Title)
		nodeJSON, err = json.MarshalIndent(node, "", "  ")
		err = ioutil.WriteFile(filepath.Join(path, "node.json"), nodeJSON, 0644)
		if err != nil {
			return fmt.Errorf("Could not write node.json: %v", err)
		}
		return nil
	}

	err := filepath.Walk(flag.Arg(0), walker)
	if err != nil {
		log.Printf("Could not walk nodes: %v", err)
		os.Exit(1)
	}
}
