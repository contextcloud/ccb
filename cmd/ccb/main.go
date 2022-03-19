package main

import (
	"fmt"
	"io"
	"os"

	"github.com/contextcloud/ccb-cli/pkg/cmd"
	"github.com/spf13/cobra"
)

type exitCode int

const (
	exitOK     exitCode = 0
	exitError  exitCode = 1
	exitCancel exitCode = 2
	exitAuth   exitCode = 4
)

func main() {
	code := mainRun()
	os.Exit(int(code))
}

func mainRun() exitCode {
	// deps := app.NewDependencies()

	stderr := os.Stderr
	rootCmd := cmd.NewRootCmd()
	if cmd, err := rootCmd.ExecuteC(); err != nil {
		printError(stderr, err, cmd)
	}
	return exitOK
}

func printError(out io.Writer, err error, cmd *cobra.Command) {
	fmt.Fprintln(out, err)
}
