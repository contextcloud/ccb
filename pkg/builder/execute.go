package builder

import (
	"fmt"
	"os"
	"os/exec"
)

// ExecCommand run a system command
func ExecCommand(path string, builder []string) error {
	targetCmd := exec.Command(builder[0], builder[1:]...)
	targetCmd.Dir = path
	targetCmd.Stdout = os.Stdout
	targetCmd.Stderr = os.Stderr
	if err := targetCmd.Start(); err != nil {
		return fmt.Errorf("ERROR - Could not start command: %v %s", builder, err)
	}
	if err := targetCmd.Wait(); err != nil {
		return fmt.Errorf("ERROR - Could not execute command: %v %w", builder, err)
	}
	return nil
}
