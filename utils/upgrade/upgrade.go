// Tool to upgrade the node files

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"pkg.monsti.org/monsti/api/service"
)

type tnodeJSON struct {
	service.Node
	Type   string
	Fields service.NestedMap
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
		var node tnodeJSON

		err = json.Unmarshal(nodeJSON, &node)
		if err != nil {
			return fmt.Errorf("Could not unmarshal node.json: %v", err)
		}
		node.PublishTime = time.Now()
		node.Changed = time.Now()
		node.Public = true
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
