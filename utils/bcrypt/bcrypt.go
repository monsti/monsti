// Tool to generate bcrypt password hashes.
package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/howeyc/gopass"
	"golang.org/x/crypto/bcrypt"
)

func error(args ...interface{}) {
	fmt.Println(args...)
	os.Exit(1)
}

func main() {
	fmt.Printf("Enter password to be hashed: ")
	password, err := gopass.GetPasswd()
	if err != nil {
		error("Could not read password:", err)
	}
	fmt.Printf("Repeat password: ")
	confirmation, err := gopass.GetPasswd()
	if err != nil {
		error("Could not read password:", err)
	}
	if bytes.Equal(password, confirmation) {
		error("Passwords do not match.")
	}
	hash, err := bcrypt.GenerateFromPassword(password, 0)
	if err != nil {
		error("Could not hash password:", err)
	}
	fmt.Printf("%s\n", hash)
}
