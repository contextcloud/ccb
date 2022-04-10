package builder

import (
	"bufio"
	"context"
	"os"
	"path"
)

type DockerArgs struct {
	Args    BuildArgs
	Network string
	Image   string
	Path    string
}

func (b *DockerArgs) Push(ctx context.Context) error {
	args := []string{"push", b.Image}
	return execCommand(".", "docker", args, nil)
}

func (b *DockerArgs) Build(ctx context.Context) error {
	args := []string{"build", "--progress=plain"}
	args = append(args, b.Args.CmdArgs()...)
	if len(b.Network) > 0 {
		args = append(args, "--network", b.Network)
	}
	args = append(args, "-t", b.Image)

	if len(path.Ext(b.Path)) > 0 {
		args = append(args, "-")

		file, err := os.Open(b.Path)
		if err != nil {
			return nil
		}
		buf := bufio.NewReader(file)
		return execCommand(".", "docker", args, buf)
	}

	args = append(args, b.Path)
	return execCommand(".", "docker", args, nil)
}
