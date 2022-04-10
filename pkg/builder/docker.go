package builder

import (
	"context"
	"path"
)

type DockerArgs struct {
	Args    BuildArgs
	Network string
	Image   string
	Path    string
}

func (b *DockerArgs) Push(ctx context.Context) error {
	args := []string{"docker", "push", b.Image}
	return ExecCommand(".", args)
}

func (b *DockerArgs) Build(ctx context.Context) error {
	args := []string{"docker", "build"}
	args = append(args, b.Args.CmdArgs()...)
	if len(b.Network) > 0 {
		args = append(args, "--network", b.Network)
	}
	args = append(args, "-t", b.Image)

	if len(path.Ext(b.Path)) > 0 {
		args = append(args, "-", "<", b.Path)
	} else {
		args = append(args, b.Path)
	}

	return ExecCommand(".", args)
}
