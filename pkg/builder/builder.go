package builder

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/contextcloud/ccb/pkg/print"
	"github.com/contextcloud/ccb/pkg/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/neilotoole/errgroup"
)

type BuildResult struct {
	Image string
}

type Build interface {
	Name() string
	Run(ctx context.Context) (*BuildResult, error)
}

// BuildArgs make prettier
type BuildArgs map[string]*string

type BuildFunction struct {
	Args    BuildArgs
	Network string
	Image   string

	Name         string
	FilesPath    string
	TemplatePath string
}

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

	cli       *client.Client
	functions []Build
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

func (b *builder) toBuild(name string, template string, args BuildArgs) (Build, error) {
	if utils.IsDockerTemplate(template) {
		return NewDockerfileBuild(b, name, args)
	}

	return NewPackBuild(b, name, template, args)
}

func (b *builder) AddService(name string, template string, args BuildArgs) error {
	build, err := b.toBuild(name, template, args)
	if err != nil {
		return err
	}
	b.functions = append(b.functions, build)
	return nil
}

func (b *builder) Build(ctx context.Context) ([]string, error) {
	cpus := runtime.NumCPU()
	g, ctx := errgroup.WithContextN(ctx, cpus, b.PoolSize)

	var out []string
	for _, doc := range b.functions {
		v := doc
		name := v.Name()
		out = append(out, name)

		g.Go(func() error {
			// run the build
			result, err := v.Run(ctx)
			if err != nil {
				return err
			}
			if result == nil || result.Image == "" {
				return errors.New("no image")
			}

			// do we push?
			if err := b.push(ctx, result.Image); err != nil {
				return err
			}

			return nil
		})

	}

	return out, g.Wait()
}

func (b *builder) push(ctx context.Context, image string) error {
	if !b.Push {
		return nil
	}

	b.Log.Printf("%s: Pushing image\n", image)

	pushOptions := types.ImagePushOptions{
		RegistryAuth: b.RegistryAuth,
	}
	pushResp, err := b.cli.ImagePush(ctx, image, pushOptions)
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
