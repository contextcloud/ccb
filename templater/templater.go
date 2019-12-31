package templater

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/go-getter"
	"golang.org/x/sync/errgroup"
)

const defaultTemplateLocation = "github.com/contextgg/openfaas-templates"

// TemplaterOption options for our templater
type TemplaterOption func(*Client)

func AddLocationOption(name, location string) func(*Client) {
	return func(c *Client) {
		c.templateLocations[name] = location
	}
}

type Templater interface {
	AddFunction(string, string)
	Download(context.Context) ([]string, error)
	Pack(context.Context) ([]string, error)
}

func NewTemplater(opts ...TemplaterOption) Templater {
	c := &Client{
		templateLocations: make(map[string]string),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type templateFunction struct {
	Name     string
	Template string
}

type Client struct {
	templateLocations map[string]string
	functions         []templateFunction
}

func (c *Client) AddFunction(name string, template string) {
	c.functions = append(c.functions, templateFunction{name, template})
}

func (c *Client) getTemplate(template string) string {
	// get the source.!
	loc, ok := c.templateLocations[template]
	if !ok {
		loc = defaultTemplateLocation
	}

	// build the location!.
	// strip the "https://"
	if strings.HasPrefix(loc, "https://") {
		loc = loc[:8]
	}

	if strings.HasSuffix(loc, "/") {
		loc = loc[0 : len(loc)-1]
	}

	return fmt.Sprintf("%s/template//%s", loc, template)
}

func (c *Client) Download(ctx context.Context) ([]string, error) {
	// build a list of functions
	templates := make(map[string]string)
	for _, fn := range c.functions {
		if _, ok := templates[fn.Template]; ok {
			continue
		}
		templates[fn.Template] = c.getTemplate(fn.Template)
	}

	// make a go channel!.
	os.RemoveAll(path.Join(".", "templates"))
	return downloadAll(ctx, templates)
}
func (c *Client) Pack(ctx context.Context) ([]string, error) {
	return nil, nil
}

type downloadJob struct {
	source   string
	template string
}
type downloadResult struct {
	source   string
	template string
}

func downloadAll(ctx context.Context, templates map[string]string) ([]string, error) {
	g, ctx := errgroup.WithContext(ctx)
	jobs := make(chan downloadJob)

	g.Go(func() error {
		defer close(jobs)
		for name, tmpl := range templates {
			select {
			case jobs <- downloadJob{tmpl, name}:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})

	c := make(chan downloadResult)

	const numDigesters = 20
	for i := 0; i < numDigesters; i++ {
		g.Go(func() error {
			for job := range jobs {
				err := download(job.source, job.template)
				if err != nil {
					return err
				}
				select {
				case c <- downloadResult{job.source, job.template}:
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
		out = append(out, r.source)
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return out, nil
}

// pullTemplate using go-getter
func download(repository string, template string) error {
	cli := &getter.Client{
		Mode: getter.ClientModeDir,
		Src:  repository,
		Dst:  fmt.Sprintf("template/%s", template),
		Pwd:  ".",
	}
	return cli.Get()
}
