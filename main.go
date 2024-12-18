package main

import (
	"os"

	"github.com/iamhectorsosa/octomap/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
