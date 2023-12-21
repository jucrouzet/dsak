// Package main parses a commit message and checks if it is valid against a conventionalcommits rules.
package main

import (
	"fmt"
	"os"

	"github.com/leodido/go-conventionalcommits"
	"github.com/leodido/go-conventionalcommits/parser"
)

func main() {
	var message []byte
	var err error
	if len(os.Args) > 1 {
		message, err = os.ReadFile(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read commit message file: %s\n", err.Error())
			os.Exit(255)
		}
	} else {
		message, err = os.ReadFile(os.Stdin.Name())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read stding: %s\n", err.Error())
			os.Exit(255)
		}
	}
	message = append(message, '\n')
	if _, err = parser.NewMachine(parser.WithTypes(conventionalcommits.TypesConventional)).Parse(message); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing commit message: %s\n", err.Error())
		os.Exit(255)
	}
	os.Exit(0)
}
