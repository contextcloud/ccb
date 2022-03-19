package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/contextcloud/ccb-cli/pkg/version"
)

// GitCommit injected at build-time
var (
	shortVersion bool
	warnUpdate   bool
)

func init() {
	versionCmd.Flags().BoolVar(&shortVersion, "short-version", false, "Just print Git SHA")

	rootCmd.AddCommand(versionCmd)
}

// versionCmd displays version information
var versionCmd = &cobra.Command{
	Use:   "version [--short-version]",
	Short: "Display the clients version information",
	Long: fmt.Sprintf(`The version command returns the current clients version information.
This currently consists of the GitSHA from which the client was built.
- https://github.com/contextcloud/ccb-cli/tree/%s`, version.GitCommit),
	Example: `  ccb version
  ccb version --short-version`,
	RunE: runVersionE,
}

func runVersionE(cmd *cobra.Command, args []string) error {
	if shortVersion {
		fmt.Println(version.BuildVersion())
	} else {
		fmt.Printf(`CLI:
 commit:  %s
 version: %s
`, version.GitCommit, version.BuildVersion())
	}

	return nil
}
