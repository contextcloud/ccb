package templater

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/go-safetemp"
	"golang.org/x/sync/errgroup"
)

const defaultTemplateLocation = "github.com/contextcloud/templates"

// Option options for our templater
type Option func(*Client)

// AddLocationOption add locations for partical templates
func AddLocationOption(name, location string) func(*Client) {
	return func(c *Client) {
		c.templateLocations[name] = location
	}
}

// Templater interface
type Templater interface {
	AddFunction(string, string)
	Download(context.Context) ([]string, error)
	Pack(context.Context) ([]string, error)
}

// NewTemplater will create a new templater
func NewTemplater(opts ...Option) Templater {
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

// Client struct
type Client struct {
	templateLocations map[string]string
	functions         []templateFunction
}

// AddFunction will add a name and template
func (c *Client) AddFunction(name string, template string) {
	c.functions = append(c.functions, templateFunction{name, template})
}

func (c *Client) getTemplate(template string) string {
	// get the source.!
	loc, ok := c.templateLocations[template]
	if !ok || len(loc) == 0 {
		loc = defaultTemplateLocation
	}

	// build the location!.
	// strip the "https://"
	loc = strings.TrimPrefix(loc, "https://")

	if strings.HasSuffix(loc, "/") {
		loc = loc[0 : len(loc)-1]
	}

	return fmt.Sprintf("%s/template//%s", loc, template)
}

// Download will fetch in parallel
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
	os.RemoveAll(path.Join(".", "template"))
	return downloadAll(ctx, templates)
}

// Pack will create buildable functions
func (c *Client) Pack(ctx context.Context) ([]string, error) {
	os.RemoveAll(path.Join(".", "build"))
	return packAll(ctx, c.functions)
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

type packResult struct {
	functionName string
	template     string
}

func packAll(ctx context.Context, templates []templateFunction) ([]string, error) {
	g, ctx := errgroup.WithContext(ctx)
	jobs := make(chan templateFunction)

	g.Go(func() error {
		defer close(jobs)
		for _, fn := range templates {
			select {
			case jobs <- fn:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})

	c := make(chan packResult)

	const numDigesters = 20
	for i := 0; i < numDigesters; i++ {
		g.Go(func() error {
			for job := range jobs {
				err := pack(job.Template, job.Name, nil)
				if err != nil {
					return err
				}
				select {
				case c <- packResult{job.Name, job.Template}:
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

func pack(templateName, fnName string, args map[string]string) error {
	ctx := context.Background()

	realDst := path.Join(".", "build", fnName)
	tmpl := path.Join(".", "template", templateName)

	td, tdcloser, err := safetemp.Dir("", "ccb-cli")
	if err != nil {
		return err
	}
	defer tdcloser.Close()

	tmplFn := path.Join(td, "function")
	// delete tmp dir
	if err := os.RemoveAll(td); err != nil {
		return fmt.Errorf("Could not delete %s %w", td, err)
	}
	// make tmp dir
	if err := os.MkdirAll(td, 0755); err != nil {
		return fmt.Errorf("Could not create temp: %s %w", td, err)
	}
	// copy the template.
	if err := copyDir(ctx, td, tmpl, false); err != nil {
		return fmt.Errorf("Could not copy template %s to temp %s: %w", tmpl, td, err)
	}
	// delete the function
	if err := os.RemoveAll(tmplFn); err != nil {
		return fmt.Errorf("Could not delete function example in temp: %w", err)
	}

	// TODO find all templates and execute them!

	// make the function dir
	if err := os.MkdirAll(tmplFn, 0755); err != nil {
		return fmt.Errorf("Could not create function dir in temp: %w", err)
	}
	if err := copyDir(ctx, tmplFn, fnName, false); err != nil {
		return fmt.Errorf("Could not copy function to temp: %w", err)
	}

	// move it!
	// delete the old
	if err := os.RemoveAll(realDst); err != nil {
		return fmt.Errorf("Could not clean pack destination: %w", err)
	}
	// make the real dir
	if err := os.MkdirAll(realDst, 0755); err != nil {
		return fmt.Errorf("Could not create pack destination: %w", err)
	}
	if err := copyDir(ctx, realDst, td, false); err != nil {
		return fmt.Errorf("Could not copy pack to dest: %w", err)
	}
	return nil
}
