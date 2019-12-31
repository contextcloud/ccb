package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// Flags that are to be added to all commands.
var (
	yamlFile string
	regex    string
	filter   string
)

func init() {
	// Setup terminal std
	rootCmd.PersistentFlags().StringVarP(&yamlFile, "yaml", "f", "", "Path to YAML file describing function(s)")
	rootCmd.PersistentFlags().StringVarP(&regex, "regex", "", "", "Regex to match with function names in YAML file")
	rootCmd.PersistentFlags().StringVarP(&filter, "filter", "", "", "Wildcard to match with function names in YAML file")

	// Set Bash completion options
	validYAMLFilenames := []string{"yaml", "yml"}
	_ = rootCmd.PersistentFlags().SetAnnotation("yaml", cobra.BashCompFilenameExt, validYAMLFilenames)
}

var rootCmd = &cobra.Command{
	Use:   "faas-cd",
	Short: "Manage your OpenFaaS functions from the command line",
	Long: `
Manage your OpenFaaS functions from the command line`,
	Run: runFaasCD,
}

func runFaasCD(cmd *cobra.Command, args []string) {
}

func Execute(customArgs []string) {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	rootCmd.SetArgs(customArgs[1:])
	if err := rootCmd.Execute(); err != nil {
		e := err.Error()
		fmt.Println(strings.ToUpper(e[:1]) + e[1:])
		os.Exit(1)
	}
}
