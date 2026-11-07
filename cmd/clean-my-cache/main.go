package main

import (
	"os"

	"github.com/jefferson-calmon/clean-my-cache/internal/cli"
)

var version = "dev"

func main() {
	os.Exit(cli.Execute(version))
}
