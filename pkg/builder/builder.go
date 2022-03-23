package builder

import (
	"context"
	"fmt"
	"path"
	"strings"

	"golang.org/x/sync/errgroup"
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

func buildImage(b *dockerBuildOpts) error {
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

type buildJob struct {
	functionName string
	buildOpts    *dockerBuildOpts
}
type buildResult struct {
	functionName string
}

func buildAll(ctx context.Context, functions map[string]*dockerBuildOpts) ([]string, error) {
	g, ctx := errgroup.WithContext(ctx)
	jobs := make(chan buildJob)

	g.Go(func() error {
		defer close(jobs)
		for name, opts := range functions {
			select {
			case jobs <- buildJob{name, opts}:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})

	c := make(chan buildResult)

	const numDigesters = 1
	for i := 0; i < numDigesters; i++ {
		g.Go(func() error {
			for job := range jobs {
				err := buildImage(job.buildOpts)
				if err != nil {
					return err
				}
				select {
				case c <- buildResult{job.functionName}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}
	go func() {
		g.Wait()
		close(c)
	}()

	var out []string
	for r := range c {
		out = append(out, r.functionName)
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return out, nil
}

func pushImage(image string) error {
	args := []string{"docker", "push", image}
	return ExecCommand(".", args)
}

type pushJob struct {
	functionName string
	image        string
}
type pushResult struct {
	functionName string
}

func pushAll(ctx context.Context, functions map[string]string) ([]string, error) {
	g, ctx := errgroup.WithContext(ctx)
	jobs := make(chan pushJob)

	g.Go(func() error {
		defer close(jobs)
		for name, image := range functions {
			select {
			case jobs <- pushJob{name, image}:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})

	c := make(chan pushResult)

	const numDigesters = 1
	for i := 0; i < numDigesters; i++ {
		g.Go(func() error {
			for job := range jobs {
				err := pushImage(job.image)
				if err != nil {
					return err
				}
				select {
				case c <- pushResult{job.functionName}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}
	go func() {
		g.Wait()
		close(c)
	}()

	var out []string
	for r := range c {
		out = append(out, r.functionName)
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return out, nil
}
