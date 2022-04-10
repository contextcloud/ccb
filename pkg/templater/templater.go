package templater

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
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
	Tar(ctx context.Context) ([]string, error)
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

// Pack will create buildable functions
func (t *templater) Pack(ctx context.Context) ([]string, error) {
	cpus := runtime.NumCPU()
	g, ctx := errgroup.WithContextN(ctx, cpus, cpus)

	var out []string
	for _, fn := range t.functions {
		out = append(out, fn.Name)
		g.Go(func() error {
			return t.pack(ctx, fn.Template, fn.Name)
		})
	}
	return out, g.Wait()
}

// Tar will create tarballs for functions
func (t *templater) Tar(ctx context.Context) ([]string, error) {
	cpus := runtime.NumCPU()
	g, ctx := errgroup.WithContextN(ctx, cpus, cpus)

	var out []string
	for _, fn := range t.functions {
		out = append(out, fn.Name)

		n := fn.Name
		v := fn.Template

		g.Go(func() error {
			if err := t.tar(ctx, v, n); err != nil {
				return err
			}
			return nil
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
		Mode: getter.ClientModeDir,
		Src:  repository,
		Dst:  fmt.Sprintf(".ccb/%s/%s", templatesDir, template),
		Pwd:  ".",
	}
	return cli.Get()
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

func (t *templater) tar(ctx context.Context, templateName, fnName string) error {
	destination := path.Join(".", ".ccb", buildDir, fnName+".tar.gz")
	templateSrc := path.Join(".", ".ccb", templatesDir, templateName)
	functionSrc := path.Join(t.workingDir, fnName)
	ignoreNames := []string{".gitignore", ".dockerignore"}

	// remove all
	if err := os.RemoveAll(destination); err != nil {
		fmt.Println("RemoveAll")
		return err
	}

	// ensure dir exists
	if err := os.MkdirAll(path.Dir(destination), 0777); err != nil {
		fmt.Println("MkdirAll")
		return err
	}

	out, err := os.Create(destination)
	if err != nil {
		fmt.Println("Create")
		return err
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	folders := map[string]string{
		templateSrc: "",
		functionSrc: "function",
	}
	for folder, out := range folders {
		var ignores []gitignore.GitIgnore
		for _, filename := range ignoreNames {
			ignorePath := path.Join(folder, filename)
			ignore, err := gitignore.NewFromFile(ignorePath)
			if err != nil && !os.IsNotExist(err) {
				fmt.Println("IsNotExist")
				return err
			}
			if ignore != nil {
				ignores = append(ignores, ignore)
			}
		}

		walkFn := func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Mode().IsDir() {
				return nil
			}

			// Because of scoping we can reference the external root_directory variable
			np := p[len(folder)+1:]
			if len(np) == 0 {
				return nil
			}

			// skip stuff
			for _, ign := range ignores {
				m := ign.Match(p)
				if m != nil && m.Ignore() {
					return nil
				}
			}

			name := np
			if len(out) > 0 {
				name = path.Join(out, np)
			}

			var link string
			if info.Mode()&os.ModeSymlink == os.ModeSymlink {
				if link, err = os.Readlink(p); err != nil {
					return err
				}
			}

			h, err := tar.FileInfoHeader(info, link)
			if err != nil {
				return err
			}
			h.Name = name
			if err := tw.WriteHeader(h); err != nil {
				return err
			}

			fr, err := os.Open(p)
			if err != nil {
				return err
			}
			defer fr.Close()

			if _, err := io.Copy(tw, fr); err != nil {
				return err
			}
			return nil
		}

		if err := filepath.Walk(folder, walkFn); err != nil {
			return err
		}
	}

	return nil
}
