// Tool to generate bcrypt password hashes.
package main

import (
	"code.google.com/p/go.crypto/bcrypt"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Println("%v does not expect any command line arguments")
		os.Exit(1)
	}
	password, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println("Could not read password from standard input: %v", err)
		os.Exit(1)
	}
	hash, err := bcrypt.GenerateFromPassword(password, 0)
	if err != nil {
		fmt.Println("Could not hash password: %v", err)
		os.Exit(1)
	}
	fmt.Printf("%s\n", hash)
}
