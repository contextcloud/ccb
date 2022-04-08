package templater

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/denormal/go-gitignore"
	"github.com/hashicorp/go-getter"
	"github.com/neilotoole/errgroup"
	cp "github.com/otiai10/copy"
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

	loc = strings.TrimPrefix(loc, "https://")
	loc = strings.TrimSuffix(loc, "/")

	return fmt.Sprintf("%s//%s", loc, template)
}

func (t *templater) downloadAll(ctx context.Context, templates map[string]string) ([]string, error) {
	cpus := runtime.NumCPU()
	g, ctx := errgroup.WithContextN(ctx, cpus, 1)

	var out []string
	for name, tmpl := range templates {
		out = append(out, name)
		g.Go(func() error {
			return t.download(ctx, tmpl, name)
		})
	}

	return out, g.Wait()
}

func (t *templater) download(ctx context.Context, repository, template string) error {
	cli := &getter.Client{
		Mode: getter.ClientModeDir,
		Src:  repository,
		Dst:  fmt.Sprintf(".ccb/%s/%s", templatesDir, template),
		Pwd:  ".",
	}
	return cli.Get()
}

func (t *templater) packAll(ctx context.Context, templates []templateFunction) ([]string, error) {
	cpus := runtime.NumCPU()
	g, ctx := errgroup.WithContextN(ctx, cpus, cpus)

	var out []string
	for _, fn := range templates {
		out = append(out, fn.Name)
		g.Go(func() error {
			return t.pack(ctx, fn.Template, fn.Name)
		})
	}
	return out, g.Wait()
}

func (t *templater) pack(ctx context.Context, templateName, fnName string) error {
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
