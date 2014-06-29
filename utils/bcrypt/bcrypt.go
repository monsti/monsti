// Tool to generate bcrypt password hashes.
package main

import (
	"fmt"
	"os"

	"code.google.com/p/go.crypto/bcrypt"
	"code.google.com/p/gopass"
)

func error(args ...interface{}) {
	fmt.Println(args...)
	os.Exit(1)
}

func main() {
	password, err := gopass.GetPass("Enter password to be hashed: ")
	if err != nil {
		error("Could not read password:", err)
	}
	confirmation, err := gopass.GetPass("Repeat password: ")
	if err != nil {
		error("Could not read password:", err)
	}
	if password != confirmation {
		error("Passwords do not match.")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		error("Could not hash password:", err)
	}
	fmt.Printf("%s\n", hash)
}
