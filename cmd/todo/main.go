package main

import (
	"fmt"
	"os"

	"github.com/frankbardon/todo/cmd/todo/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
