// Tool to generate bcrypt password hashes.
package main

import (
	"code.google.com/p/go.crypto/bcrypt"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Printf("Usage: %v <password>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(flag.Arg(0)), 0)
	if err != nil {
		fmt.Println("Could not hash password: %v", err)
		os.Exit(1)
	}
	fmt.Printf("%s\n", hash)
}
