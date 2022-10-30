package templater

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/hashicorp/go-getter"
	"github.com/neilotoole/errgroup"
)

const defaultTemplateLocation = "github.com/contextcloud/templates"
const templatesDir = "templates"
const buildDir = "build"
const functionDir = "function"

type templateFunction struct {
	Name     string
	Template string
}

// Templater interface
type Templater interface {
	AddFunction(name string, template string)
	Download(ctx context.Context) ([]string, error)
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

	cpus := runtime.NumCPU()
	g, ctx := errgroup.WithContextN(ctx, cpus, 1)

	var out []string
	for name, tmpl := range templates {
		out = append(out, name)

		n := name
		v := tmpl

		g.Go(func() error {
			return t.download(ctx, v, n)
		})
	}

	return out, g.Wait()
}

func (t *templater) getTemplate(template string) string {
	// get the source.!
	loc, ok := t.templateLocations[template]
	if !ok || len(loc) == 0 {
		loc = defaultTemplateLocation
	}

	loc = strings.TrimPrefix(loc, "https://")
	loc = strings.TrimSuffix(loc, "/")

	return fmt.Sprintf("%s//%s", loc, template)
}

func (t *templater) download(ctx context.Context, repository, template string) error {
	cli := &getter.Client{
		Ctx:  ctx,
		Mode: getter.ClientModeDir,
		Src:  repository,
		Dst:  fmt.Sprintf(".ccb/%s/%s", templatesDir, template),
		Pwd:  ".",
	}
	return cli.Get()
}
