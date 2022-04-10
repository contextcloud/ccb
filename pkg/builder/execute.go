package builder

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

func execCommand(path string, name string, args []string, input io.Reader) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = path
	cmd.Stdin = input
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ERROR - Could not start command: %v %s", name, err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ERROR - Could not execute command: %v %w", name, err)
	}
	return nil
}
