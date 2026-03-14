package main

import (
	"github.com/henryppercy/hp-source/cmd"
	"os"
)

func main() {
	cmd.Run(os.Args[1:])
}
