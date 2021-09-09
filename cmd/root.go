package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const (
	defaultYAML = "stack.yml"
)

// Flags that are to be added to all commands.
var (
	yamlFile  string
	regex     string
	filter    string
	network   string
	buildArgs []string
)

func init() {
	// Setup terminal std
	rootCmd.PersistentFlags().StringVarP(&yamlFile, "yaml", "f", "", "Path to YAML file describing function(s)")
	rootCmd.PersistentFlags().StringVarP(&regex, "regex", "", "", "Regex to match with function names in YAML file")
	rootCmd.PersistentFlags().StringVarP(&filter, "filter", "", "", "Wildcard to match with function names in YAML file")

	rootCmd.PersistentFlags().StringVarP(&network, "network", "", "", "The network to connect to")
	rootCmd.PersistentFlags().StringSliceVarP(&buildArgs, "build-args", "", []string{}, "Wildcard to match with function names in YAML file")

	// Set Bash completion options
	validYAMLFilenames := []string{"yaml", "yml"}
	_ = rootCmd.PersistentFlags().SetAnnotation("yaml", cobra.BashCompFilenameExt, validYAMLFilenames)
}

var rootCmd = &cobra.Command{
	Use:   "ccb",
	Short: "Manage your Context Cloud functions from the command line",
	Long: `
Manage your Context Cloud functions from the command line`,
	RunE: runFaasCD,
}

func runFaasCD(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

func checkAndSetDefaultYaml() {
	// Check if there is a default yaml file and set it
	if _, err := os.Stat(defaultYAML); err == nil {
		yamlFile = defaultYAML
	}
}

func Execute(customArgs []string) {
	checkAndSetDefaultYaml()

	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	rootCmd.SetArgs(customArgs[1:])
	if err := rootCmd.Execute(); err != nil {
		e := err.Error()
		fmt.Println(strings.ToUpper(e[:1]) + e[1:])
		os.Exit(1)
	}
}
