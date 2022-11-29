package builder

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/contextcloud/ccb/pkg/builder/resources"
	"github.com/contextcloud/ccb/pkg/utils"
	"github.com/docker/docker/api/types"
)

type packBuild struct {
	builder *builder

	name         string
	buildArgs    BuildArgs
	filesPath    string
	templatePath string
}

func (d *packBuild) Name() string {
	return d.name
}

func (d *packBuild) function(ctx context.Context) (string, error) {
	// build the first one!
	d.builder.Log.Printf("%s: Building files\n", d.name)

	dir := NewDirArchive(d.filesPath, true)
	dockerfile := NewRawArchive("Dockerfile", resources.FilesDockerFile)

	reader, err := buildArchive(dir, dockerfile)
	if err != nil {
		return "", err
	}

	buildOptions := types.ImageBuildOptions{
		Context:    reader,
		Dockerfile: "Dockerfile",
		Remove:     true,
		BuildArgs: map[string]*string{
			"FILES": &dir.Name,
		},
	}

	imageResp, err := d.builder.cli.ImageBuild(ctx, reader, buildOptions)
	if err != nil {
		return "", err
	}
	defer imageResp.Body.Close()

	// parse the output
	auxs, err := buildResult(imageResp.Body, d.builder.Log)
	if err != nil {
		return "", err
	}

	if len(auxs) == 0 {
		return "", errors.New("no aux data")
	}

	return auxs[len(auxs)-1].Id, nil
}

func (d *packBuild) handler(ctx context.Context, filesImage string) (string, error) {
	// build the first one!
	d.builder.Log.Printf("%s: Building function\n", d.name)

	template := NewDirArchive(d.templatePath, true)
	reader, err := buildArchive(template)
	if err != nil {
		return "", err
	}

	hargs := map[string]*string{
		"FUNCTION_IMG": &filesImage,
	}
	args := utils.MergeMap(d.buildArgs, hargs)

	imageName := d.builder.imageName(d.name)
	buildOptions := types.ImageBuildOptions{
		Context:    reader,
		Dockerfile: fmt.Sprintf("%s/Dockerfile", template.Name),
		Remove:     true,
		Tags:       []string{imageName},
		BuildArgs:  args,
	}

	imageResp, err := d.builder.cli.ImageBuild(ctx, reader, buildOptions)
	if err != nil {
		return "", err
	}
	defer imageResp.Body.Close()

	// parse the output
	auxs, err := buildResult(imageResp.Body, d.builder.Log)
	if err != nil {
		return "", err
	}

	if len(auxs) == 0 {
		return "", errors.New("no aux data")
	}

	return imageName, nil
}

func (d *packBuild) Run(ctx context.Context) (*BuildResult, error) {
	// what's the files?
	functionImg, err := d.function(ctx)
	if err != nil {
		return nil, err
	}

	handlerImg, err := d.handler(ctx, functionImg)
	if err != nil {
		return nil, err
	}

	return &BuildResult{
		Image: handlerImg,
	}, nil
}

func NewPackBuild(builder *builder, name string, template string, buildArgs BuildArgs) (Build, error) {
	tpath := path.Join(".", ".ccb", "templates", template)
	fpath := path.Join(builder.WorkingDir, name)

	// check if the template exists
	if _, err := os.Stat(tpath); err != nil {
		return nil, err
	}
	// check if the files exists
	if _, err := os.Stat(fpath); err != nil {
		return nil, err
	}

	return &packBuild{
		builder:      builder,
		name:         name,
		buildArgs:    buildArgs,
		filesPath:    fpath,
		templatePath: tpath,
	}, nil
}
