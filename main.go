package main

import (
	"fmt"
	"os"

	"github.com/mattn/go-colorable"
	"github.com/mkideal/cli"
)

func main() {
	cli.SetUsageStyle(cli.NormalStyle)
	if err := rootCommand.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(colorable.NewColorableStderr(), err)
		os.Exit(1)
	}
}
