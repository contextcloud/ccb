package deployer

import (
	"bytes"
	"errors"
	"os"
	"text/template"

	rice "github.com/GeertJohan/go.rice"
	"github.com/contextcloud/ccb-cli/pkg/deployer/templates"
	"github.com/contextcloud/ccb-cli/pkg/utils"
)

var (
	// ErrNoMetadata when a file does not contain metadata
	ErrNoMetadata = errors.New("No metadata found in file")
	// ErrNoSpec when a file does not contain spec
	ErrNoSpec = errors.New("No spec found in file")
	// ErrInvalidKind kind is not support
	ErrInvalidKind = errors.New("Unsupported kind")
	// ErrInvalidNamespace when two namespaces don't match
	ErrInvalidNamespace = errors.New("Namespaces don't match")
	// ErrNoConfig when the config isn't supplied
	ErrNoConfig = errors.New("No config supplied")
)

var livenessProbe = &Probe{
	Enabled:             true,
	Path:                "/live",
	Port:                "health",
	InitialDelaySeconds: 5,
	TimeoutSeconds:      5,
	PeriodSeconds:       5,
}
var readinessProbe = &Probe{
	Enabled:             true,
	Path:                "/ready",
	Port:                "health",
	InitialDelaySeconds: 5,
	TimeoutSeconds:      5,
	PeriodSeconds:       5,
}

type Manager interface {
	AddFunction(fn *Function)
	Generate() (Manifests, error)
}

type manager struct {
	box        *rice.Box
	workingDir string
	namespace  string
	registry   string
	tag        string
	functions  []*Function
}

func (m *manager) mergeEnv(all map[string]Environment, files []string, env map[string]string) (map[string]string, error) {
	var out Environment
	for _, name := range files {
		filename, err := utils.YamlFile(m.workingDir, name)
		if err != nil {
			return nil, err
		}

		found, ok := all[filename]
		if !ok {
			return nil, ErrNoConfig
		}

		out = utils.MergeMap(out, found)
	}

	out = utils.MergeMap(out, env)
	return out, nil
}

func (m *manager) secretNames(all map[string]*Secret, files []string) ([]string, error) {
	var out []string
	for _, name := range files {
		filename, err := utils.YamlFile(m.workingDir, name)
		if err != nil {
			return nil, err
		}

		found, ok := all[filename]
		if !ok {
			return nil, ErrNoConfig
		}

		out = append(out, found.Name)
	}

	return out, nil
}

func (m *manager) executeFunction(dir string, key string, data map[string]interface{}) ([]Manifest, error) {
	var out []Manifest

	if err := m.box.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// Guard
		if err != nil {
			return err
		}
		// If it's a dir skip
		if info.IsDir() {
			return nil
		}

		tmplString, err := m.box.String(path)
		if err != nil {
			return err
		}

		// parse the file
		tmpl, err := template.New(path).
			Funcs(templates.FuncMap()).
			Parse(tmplString)
		if err != nil {
			return err
		}

		var tpl bytes.Buffer
		if err := tmpl.Execute(&tpl, data); err != nil {
			return err
		}

		out = append(out, Manifest{
			Type:    ToManifestType(path),
			Key:     key,
			Content: tpl.String(),
		})
		return nil
	}); err != nil {
		return nil, err
	}

	return out, nil
}

func (m *manager) AddFunction(fn *Function) {
	m.functions = append(m.functions, fn)
}

func (m *manager) Generate() (Manifests, error) {
	var all Manifests

	secrets := make(map[string]*Secret)
	envs := make(map[string]Environment)

	// load up the secrets and environments
	for _, fn := range m.functions {
		for _, secret := range fn.Secrets {
			// get the path
			name, err := utils.YamlFile(m.workingDir, secret)
			if err != nil {
				return nil, err
			}

			// already loaded
			if _, ok := secrets[name]; ok {
				continue
			}

			// load up the secrets
			secret, err := LoadSecret(name)
			if err != nil {
				return nil, err
			}

			// check the namespace
			if len(secret.Namespace) > 0 && secret.Namespace != m.namespace {
				return nil, ErrInvalidNamespace
			}

			secrets[name] = secret
		}
		for _, env := range fn.Envs {
			// get the path
			name, err := utils.YamlFile(m.workingDir, env)
			if err != nil {
				return nil, err
			}

			// already loaded
			if _, ok := envs[name]; ok {
				continue
			}

			env, err := LoadEnv(name)
			if err != nil {
				return nil, err
			}

			envs[name] = env
		}
	}

	for _, fn := range m.functions {
		// get environment files.
		env, err := m.mergeEnv(envs, fn.Envs, fn.Env)
		if err != nil {
			return nil, err
		}

		secrets, err := m.secretNames(secrets, fn.Secrets)
		if err != nil {
			return nil, err
		}

		data := map[string]interface{}{
			"Key":            fn.Key,
			"Name":           fn.Name,
			"Version":        fn.Version,
			"Namespace":      m.namespace,
			"Image":          ImageName(m.registry, fn.Key, m.tag),
			"LivenessProbe":  livenessProbe,
			"ReadinessProbe": readinessProbe,
			"Environment":    env,
			"Secrets":        secrets,
		}
		out, err := m.executeFunction("deployment", fn.Key, data)
		if err != nil {
			return nil, err
		}

		all = append(all, out...)
	}

	for _, secret := range secrets {
		all = append(all, Manifest{
			Type:    SecretManifestType,
			Key:     secret.Name,
			Content: string(secret.Raw),
		})
	}

	return all, nil
}

func NewManager(workingDir string, namespace string, registry string, tag string) Manager {
	box := templates.NewBox()

	return &manager{
		box:        box,
		workingDir: workingDir,
		namespace:  namespace,
		registry:   registry,
		tag:        tag,
	}
}
