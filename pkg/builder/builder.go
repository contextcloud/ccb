package builder

import (
	"context"
	"fmt"
	"path"
	"runtime"
	"strings"

	"github.com/neilotoole/errgroup"
)

// Option for configing the builder
type Option func(c *Client)

// SetRegistry will sent the registry which is used to build the Docker Image Name
func SetRegistry(registry string) Option {
	return func(c *Client) {
		c.registry = registry
	}
}

// SetTag used part of the Docker Image Name
func SetTag(tag string) Option {
	return func(c *Client) {
		c.tag = tag
	}
}

// SetNetwork use this network when building
func SetNetwork(network string) Option {
	return func(c *Client) {
		c.network = network
	}
}

// BuildArgs make prettier
type BuildArgs map[string]string

// CmdArgs run this into nice pretty command line args
func (b BuildArgs) CmdArgs() []string {
	var args []string
	for k, v := range b {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}
	return args
}

// Builder for building stuff.
type Builder interface {
	AddService(string, BuildArgs)
	Build(context.Context) ([]string, error)
	Push(context.Context) ([]string, error)
}

func NewBuilder(opts ...Option) Builder {
	c := &Client{
		tag:       "latest",
		functions: make(map[string]BuildArgs),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Client for building stuff in parallel
type Client struct {
	registry  string
	tag       string
	network   string
	functions map[string]BuildArgs
}

func (c *Client) AddService(fnName string, args BuildArgs) {
	c.functions[fnName] = args
}

func (c *Client) imageName(name string) string {
	if len(c.registry) == 0 {
		return fmt.Sprintf("%s:%s", name, c.tag)
	}
	if strings.HasSuffix(c.registry, "/") {
		return fmt.Sprintf("%s/%s:%s", c.registry[:len(c.registry)-1], name, c.tag)
	}
	return fmt.Sprintf("%s/%s:%s", c.registry, name, c.tag)
}

func (c *Client) Build(ctx context.Context) ([]string, error) {
	basePath := path.Join(".", ".ccb", "build")

	// build a list of functions
	builds := make(map[string]*dockerBuildOpts)
	for fn, args := range c.functions {
		// what's the image?
		image := c.imageName(fn)

		// what's the dir?
		dir := path.Join(basePath, fn)

		builds[fn] = &dockerBuildOpts{
			Image:   image,
			Dir:     dir,
			Args:    args,
			Network: c.network,
		}
	}

	return buildAll(ctx, builds)
}

func (c *Client) Push(ctx context.Context) ([]string, error) {
	// build a list of functions
	images := make(map[string]string)
	for fn := range c.functions {
		// what's the image?
		image := c.imageName(fn)
		images[fn] = image
	}

	return pushAll(ctx, images)
}

func buildImage(ctx context.Context, b *dockerBuildOpts) error {
	return ExecCommand(".", b.CmdArgs())
}

type dockerBuildOpts struct {
	Image   string
	Dir     string
	Args    BuildArgs
	Network string
}

// CmdArgs run this into nice pretty command line args
func (b *dockerBuildOpts) CmdArgs() []string {
	args := []string{"docker", "build"}
	args = append(args, b.Args.CmdArgs()...)
	args = append(args, "-t", b.Image, b.Dir)
	if len(b.Network) > 0 {
		args = append(args, "--network", b.Network)
	}
	return args
}

func buildAll(ctx context.Context, functions map[string]*dockerBuildOpts) ([]string, error) {
	cpus := runtime.NumCPU()
	g, ctx := errgroup.WithContextN(ctx, cpus, 1)

	var out []string
	for name, opts := range functions {
		out = append(out, name)
		g.Go(func() error {
			return buildImage(ctx, opts)
		})
	}

	return out, g.Wait()
}

func pushImage(ctx context.Context, image string) error {
	args := []string{"docker", "push", image}
	return ExecCommand(".", args)
}

func pushAll(ctx context.Context, functions map[string]string) ([]string, error) {
	cpus := runtime.NumCPU()
	g, ctx := errgroup.WithContextN(ctx, cpus, 1)

	var out []string
	for name, image := range functions {
		out = append(out, name)
		g.Go(func() error {
			return pushImage(ctx, image)
		})
	}

	return out, g.Wait()
}
