package builder

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/neilotoole/errgroup"
)

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
	AddService(string, BuildArgs) error
	Build(context.Context) ([]string, error)
	Push(context.Context) ([]string, error)
}

func NewBuilder(opts ...Option) Builder {
	c := &Client{
		tag:       "latest",
		functions: make(map[string]*DockerArgs),
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
	poolSize  int
	functions map[string]*DockerArgs
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

func (c *Client) AddService(fnName string, args BuildArgs) error {
	p := path.Join(".", ".ccb", "build", fnName)
	paths := []string{
		p + ".tar.gz",
		p,
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			image := c.imageName(fnName)
			c.functions[fnName] = &DockerArgs{
				Image:   image,
				Path:    p,
				Network: c.network,
				Args:    args,
			}
			return nil
		}
	}

	return fmt.Errorf("nothing to build")
}

func (c *Client) Build(ctx context.Context) ([]string, error) {
	cpus := runtime.NumCPU()
	g, ctx := errgroup.WithContextN(ctx, cpus, c.poolSize)

	var out []string
	for name, doc := range c.functions {
		out = append(out, name)

		v := doc

		g.Go(func() error {
			return v.Build(ctx)
		})
	}

	return out, g.Wait()
}

func (c *Client) Push(ctx context.Context) ([]string, error) {
	cpus := runtime.NumCPU()
	g, ctx := errgroup.WithContextN(ctx, cpus, c.poolSize)

	var out []string
	for name, doc := range c.functions {
		out = append(out, name)

		v := doc

		g.Go(func() error {
			return v.Push(ctx)
		})
	}

	return out, g.Wait()
}
