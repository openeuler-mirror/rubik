// Package main provide main function for rubik.
package main

import (
	"os"

	"isula.org/rubik/pkg/rubik"
)

func main() {
	os.Exit(rubik.Run())
}
