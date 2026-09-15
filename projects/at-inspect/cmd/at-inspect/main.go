package main

import (
	"os"

	"github.com/dayanchm/at-inspect/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
