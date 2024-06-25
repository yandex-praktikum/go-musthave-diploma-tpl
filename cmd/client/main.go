package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password, _ := bcrypt.GenerateFromPassword([]byte("password"), 8)
	fmt.Println(string(password))
}
