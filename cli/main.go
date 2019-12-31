package main

import (
	"os"

	"github.com/contextgg/faas-cd/cmd"
)

func main() {
	cmd.Execute(os.Args)
}
