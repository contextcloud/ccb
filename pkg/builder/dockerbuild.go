package builder

import (
	"context"
	"errors"
	"os"
	"path"

	"github.com/docker/docker/api/types"
)

type dockerfileBuild struct {
	builder    *builder
	name       string
	buildArgs  BuildArgs
	filesPath  string
	dockerPath string
}

func (d *dockerfileBuild) Name() string {
	return d.name
}

func (d *dockerfileBuild) Run(ctx context.Context) (*BuildResult, error) {
	// build the first one!
	d.builder.Log.Printf("%s: Building function\n", d.name)

	files := NewDirArchive(d.filesPath, false)
	reader, err := buildArchive(files)
	if err != nil {
		return nil, err
	}

	imageName := d.builder.imageName(d.name)
	buildOptions := types.ImageBuildOptions{
		Context:     reader,
		Dockerfile:  "Dockerfile",
		Remove:      true,
		Tags:        []string{imageName},
		BuildArgs:   d.buildArgs,
		NetworkMode: d.builder.Network,
	}

	imageResp, err := d.builder.cli.ImageBuild(ctx, reader, buildOptions)
	if err != nil {
		return nil, err
	}
	defer imageResp.Body.Close()

	// parse the output
	auxs, err := buildResult(imageResp.Body, d.builder.Log)
	if err != nil {
		return nil, err
	}

	if len(auxs) == 0 {
		return nil, errors.New("no aux data")
	}

	return &BuildResult{
		Image: imageName,
	}, nil
}

func NewDockerfileBuild(builder *builder, name string, args BuildArgs) (Build, error) {
	fpath := path.Join(builder.WorkingDir, name)
	dpath := path.Join(builder.WorkingDir, name, "Dockerfile")

	// check if the files exists
	if _, err := os.Stat(fpath); err != nil {
		return nil, err
	}
	// check if the files exists
	if _, err := os.Stat(dpath); err != nil {
		return nil, err
	}

	return &dockerfileBuild{
		builder:    builder,
		name:       name,
		buildArgs:  args,
		filesPath:  fpath,
		dockerPath: dpath,
	}, nil
}
