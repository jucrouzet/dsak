package main

import (
	"os"

	_ "github.com/jucrouzet/dsak/cmd"
	"github.com/jucrouzet/dsak/internal/pkg/commander"
)

func main() {
	if err := commander.Run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
