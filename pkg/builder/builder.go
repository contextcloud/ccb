package builder

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/contextcloud/ccb/pkg/builder/resources"
	"github.com/contextcloud/ccb/pkg/print"
	"github.com/contextcloud/ccb/pkg/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/neilotoole/errgroup"
)

// BuildArgs make prettier
type BuildArgs map[string]*string

type Options struct {
	Log        print.Log
	WorkingDir string
	PoolSize   int
	Network    string

	Push         bool
	Registry     string
	Prefix       string
	Tag          string
	RegistryAuth string
}

type buildFunction struct {
	Args    BuildArgs
	Network string
	Image   string

	Name         string
	FilesPath    string
	TemplatePath string
}

// Builder for building stuff.
type Builder interface {
	AddService(name string, template string, args BuildArgs) error
	Build(ctx context.Context) ([]string, error)
}

func NewBuilder(opts *Options) (Builder, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	c := &builder{
		Options: opts,
		cli:     cli,
	}
	return c, nil
}

// Client for building stuff in parallel
type builder struct {
	*Options

	functions []buildFunction
	cli       *client.Client
}

func (b *builder) imageName(name string) string {
	n := b.Prefix + name

	if len(b.Registry) == 0 {
		return fmt.Sprintf("%s:%s", n, b.Tag)
	}
	if strings.HasSuffix(b.Registry, "/") {
		return fmt.Sprintf("%s/%s:%s", b.Registry[:len(b.Registry)-1], n, b.Tag)
	}
	return fmt.Sprintf("%s/%s:%s", b.Registry, n, b.Tag)
}

func (b *builder) AddService(name string, template string, args BuildArgs) error {
	tpath := path.Join(".", ".ccb", "templates", template)
	fpath := path.Join(b.WorkingDir, name)

	// check if the template exists
	if _, err := os.Stat(tpath); err != nil {
		return err
	}
	// check if the files exists
	if _, err := os.Stat(fpath); err != nil {
		return err
	}

	image := b.imageName(name)
	bf := buildFunction{
		Network: b.Network,
		Image:   image,
		Args:    args,

		Name:         name,
		FilesPath:    fpath,
		TemplatePath: tpath,
	}
	b.functions = append(b.functions, bf)

	return nil
}

func (b *builder) Build(ctx context.Context) ([]string, error) {
	cpus := runtime.NumCPU()
	g, ctx := errgroup.WithContextN(ctx, cpus, b.PoolSize)

	var out []string
	for _, doc := range b.functions {
		v := doc
		out = append(out, v.Name)

		g.Go(func() error {
			return b.build(ctx, v)
		})
	}

	return out, g.Wait()
}

func (b *builder) function(ctx context.Context, bf buildFunction) (string, error) {
	// build the first one!
	b.Log.Printf("%s: Building files\n", bf.Name)

	dir := NewDirArchive(bf.FilesPath)
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

	imageResp, err := b.cli.ImageBuild(ctx, reader, buildOptions)
	if err != nil {
		return "", err
	}
	defer imageResp.Body.Close()

	// parse the output
	auxs, err := buildResult(imageResp.Body, b.Log)
	if err != nil {
		return "", err
	}

	if len(auxs) == 0 {
		return "", errors.New("no aux data")
	}

	return auxs[len(auxs)-1].Id, nil
}

func (b *builder) handler(ctx context.Context, bf buildFunction, filesImage string) (string, error) {
	// build the first one!
	b.Log.Printf("%s: Building function\n", bf.Name)

	template := NewDirArchive(bf.TemplatePath)
	reader, err := buildArchive(template)
	if err != nil {
		return "", err
	}

	hargs := map[string]*string{
		"FUNCTION_IMG": &filesImage,
	}
	args := utils.MergeMap(bf.Args, hargs)

	buildOptions := types.ImageBuildOptions{
		Context:    reader,
		Dockerfile: fmt.Sprintf("%s/Dockerfile", template.Name),
		Remove:     true,
		Tags:       []string{bf.Image},
		BuildArgs:  args,
	}

	imageResp, err := b.cli.ImageBuild(ctx, reader, buildOptions)
	if err != nil {
		return "", err
	}
	defer imageResp.Body.Close()

	// parse the output
	auxs, err := buildResult(imageResp.Body, b.Log)
	if err != nil {
		return "", err
	}

	if len(auxs) == 0 {
		return "", errors.New("no aux data")
	}

	return bf.Image, nil
}

func (b *builder) push(ctx context.Context, bf buildFunction, functionImage string) error {
	b.Log.Printf("%s: Pushing image\n", bf.Name)

	pushOptions := types.ImagePushOptions{
		RegistryAuth: b.RegistryAuth,
	}
	pushResp, err := b.cli.ImagePush(ctx, functionImage, pushOptions)
	if err != nil {
		return err
	}
	defer pushResp.Close()

	// parse the output
	_, err = pushResult(pushResp, b.Log)
	if err != nil {
		return err
	}

	return nil
}

func (b *builder) build(ctx context.Context, bf buildFunction) error {
	// what's the files?
	functionImg, err := b.function(ctx, bf)
	if err != nil {
		return err
	}

	handlerImg, err := b.handler(ctx, bf, functionImg)
	if err != nil {
		return err
	}

	if b.Push {
		if err := b.push(ctx, bf, handlerImg); err != nil {
			return err
		}
	}

	return nil
}
