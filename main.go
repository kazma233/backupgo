package main

import (
	"backupgo/cmd"
	"fmt"
	"os"
)

func main() {
	args := []string{"backupgo"}
	args = append(args, os.Args[1:]...)
	if err := cmd.Run(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
