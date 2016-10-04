package main

import (
	"fmt"
	"os"

	"github.com/mattn/go-colorable"
	"github.com/mkideal/cli"
	"github.com/mkideal/onepw/command"
)

func main() {
	cli.SetUsageStyle(cli.ManualStyle)
	if err := command.Exec(os.Args[1:]); err != nil {
		fmt.Fprintln(colorable.NewColorableStderr(), err)
		os.Exit(1)
	}
}
