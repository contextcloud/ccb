package templater

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/denormal/go-gitignore"
	"github.com/hashicorp/go-getter"
	cp "github.com/otiai10/copy"
	"golang.org/x/sync/errgroup"
)

const defaultTemplateLocation = "github.com/contextcloud/templates"
const templatesDir = "templates"
const buildDir = "build"
const functionDir = "function"

type templateFunction struct {
	Name     string
	Template string
}
type downloadJob struct {
	source   string
	template string
}
type downloadResult struct {
	source   string
	template string
}
type packResult struct {
	functionName string
	template     string
}

// Templater interface
type Templater interface {
	AddFunction(name string, template string)
	Download(ctx context.Context) ([]string, error)
	Pack(ctx context.Context) ([]string, error)
}

// NewTemplater will create a new templater
func NewTemplater(workingDir string) Templater {
	c := &templater{
		workingDir:        workingDir,
		templateLocations: make(map[string]string),
	}

	return c
}

type templater struct {
	workingDir        string
	templateLocations map[string]string
	functions         []templateFunction
}

// AddFunction will add a name and template
func (t *templater) AddFunction(name, template string) {
	t.functions = append(t.functions, templateFunction{name, template})
}

// Download will fetch in parallel
func (t *templater) Download(ctx context.Context) ([]string, error) {
	// build a list of functions
	templates := make(map[string]string)
	for _, fn := range t.functions {
		if _, ok := templates[fn.Template]; ok {
			continue
		}
		templates[fn.Template] = t.getTemplate(fn.Template)
	}

	// make a go channel!.
	return t.downloadAll(ctx, templates)
}

// Pack will create buildable functions
func (t *templater) Pack(ctx context.Context) ([]string, error) {
	return t.packAll(ctx, t.functions)
}

func (t *templater) getTemplate(template string) string {
	// get the source.!
	loc, ok := t.templateLocations[template]
	if !ok || len(loc) == 0 {
		loc = defaultTemplateLocation
	}

	// build the location!.
	// strip the "https://"
	loc = strings.TrimPrefix(loc, "https://")

	if strings.HasSuffix(loc, "/") {
		loc = loc[0 : len(loc)-1]
	}

	return fmt.Sprintf("%s//%s", loc, template)
}

func (t *templater) downloadAll(ctx context.Context, templates map[string]string) ([]string, error) {
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
				err := t.download(job.source, job.template)
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

func (t *templater) download(repository, template string) error {
	cli := &getter.Client{
		Mode: getter.ClientModeDir,
		Src:  repository,
		Dst:  fmt.Sprintf(".ccb/%s/%s", templatesDir, template),
		Pwd:  ".",
	}
	return cli.Get()
}

func (t *templater) packAll(ctx context.Context, templates []templateFunction) ([]string, error) {
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
				err := t.pack(job.Template, job.Name)
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

func (t *templater) pack(templateName, fnName string) error {
	destination := path.Join(".", ".ccb", buildDir, fnName)
	functionDest := path.Join(destination, functionDir)
	templateSrc := path.Join(".", ".ccb", templatesDir, templateName)
	functionSrc := path.Join(t.workingDir, fnName)
	templateIgnore := path.Join(templateSrc, functionDir)
	functionIgnore := path.Join(functionSrc, ".gitignore")

	// match a file against a particular .gitignore
	ignore, err := gitignore.NewFromFile(functionIgnore)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	var skip func(src string) (bool, error)
	if ignore != nil {
		skip = func(src string) (bool, error) {
			m := ignore.Match(src)
			return m != nil && m.Ignore(), nil
		}
	}

	// remove all
	if err := os.RemoveAll(destination); err != nil {
		return err
	}

	if err := cp.Copy(templateSrc, destination, cp.Options{
		Skip: func(src string) (bool, error) {
			return strings.EqualFold(src, templateIgnore), nil
		},
	}); err != nil {
		return err
	}

	if err := cp.Copy(functionSrc, functionDest, cp.Options{
		Skip: skip,
	}); err != nil {
		return err
	}

	return nil
}
