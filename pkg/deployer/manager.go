package deployer

import (
	"bytes"
	"embed"
	"errors"
	"io/fs"
	"strings"
	"text/template"

	"github.com/contextcloud/ccb/pkg/parser"
	"github.com/contextcloud/ccb/pkg/utils"
)

var (
	// ErrNoMetadata when a file does not contain metadata
	ErrNoMetadata = errors.New("no metadata found in file")
	// ErrNoSpec when a file does not contain spec
	ErrNoSpec = errors.New("no spec found in file")
	// ErrInvalidKind kind is not support
	ErrInvalidKind = errors.New("unsupported kind")
	// ErrInvalidNamespace when two namespaces don't match
	ErrInvalidNamespace = errors.New("namespaces don't match")
	// ErrNoConfig when the config isn't supplied
	ErrNoConfig = errors.New("no config supplied")
	// ErrInvalidFQDN when the FQDN is invalid
	ErrInvalidFQDN = errors.New("invalid FQDN")
)

//go:embed templates/*
var res embed.FS

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
	GenerateRoutes(routes []*parser.Route) (Manifests, error)
	GenerateFunctions(registry string, tag string, fn []*parser.Function) (Manifests, error)
}

type manager struct {
	workingDir string
	namespace  string
	commit     string
	funcMap    template.FuncMap
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

		out = append(out, found.SecretNames...)
	}

	return out, nil
}

func (m *manager) executeFunction(dir string, key string, data map[string]interface{}) ([]Manifest, error) {
	var out []Manifest

	err := fs.WalkDir(res, "templates/"+dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		b, err := fs.ReadFile(res, path)
		if err != nil {
			return err
		}
		tmpl, err := template.New(path).
			Funcs(m.funcMap).
			Parse(string(b))
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
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (m *manager) GenerateRoutes(routes []*parser.Route) (Manifests, error) {
	var all Manifests

	for _, r := range routes {
		data := map[string]interface{}{
			"Key":       r.Key,
			"Namespace": m.namespace,
			"Commit":    m.commit,
			"FQDN":      r.FQDN,
			"Routes":    r.Routes,
		}
		out, err := m.executeFunction("routes", "server", data)
		if err != nil {
			return nil, err
		}
		all = append(all, out...)
	}

	return all, nil
}

func (m *manager) GenerateFunctions(registry string, tag string, fns []*parser.Function) (Manifests, error) {
	var all Manifests

	routes := make(map[string][]FunctionRoute)
	secrets := make(map[string]*Secret)
	envs := make(map[string]Environment)

	// load up the secrets and environments
	for _, fn := range fns {
		for _, r := range fn.Routes {
			routes[r.Name] = append(routes[r.Name], FunctionRoute{
				Key:   fn.Key,
				Route: r,
			})
		}
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

	for _, fn := range fns {
		// get environment files.
		env, err := m.mergeEnv(envs, fn.Envs, fn.Env)
		if err != nil {
			return nil, err
		}

		// add the service envs
		env["SERVICENAME"] = fn.Key
		env["VERSION"] = fn.Version
		env["ENVIRONMENT"] = fn.Environment

		secrets, err := m.secretNames(secrets, fn.Secrets)
		if err != nil {
			return nil, err
		}

		minReplicas := 2
		maxReplicas := 4
		if fn.Replicas != nil && *fn.Replicas > 2 {
			maxReplicas = *fn.Replicas
		}

		var resources = &Resources{
			Requests: &ResourceValues{
				CPU:    "125m",
				Memory: "256Mi",
			},
			Limits: &ResourceValues{
				CPU:    "250m",
				Memory: "512Mi",
			},
		}

		if fn.Limits != nil {
			resources.Requests.CPU = fn.Limits.CPU
			resources.Requests.Memory = fn.Limits.Memory
			resources.Limits.CPU = fn.Limits.CPU
			resources.Limits.Memory = fn.Limits.Memory
		}
		if fn.Requests != nil {
			resources.Requests.CPU = fn.Requests.CPU
			resources.Requests.Memory = fn.Requests.Memory
		}

		data := map[string]interface{}{
			"Key":             fn.Key,
			"Name":            fn.Name,
			"Version":         fn.Version,
			"EnvironmentName": fn.Environment,
			"Namespace":       m.namespace,
			"Commit":          m.commit,
			"Image":           ImageName(registry, fn.Key, tag),
			"LivenessProbe":   livenessProbe,
			"ReadinessProbe":  readinessProbe,
			"Environment":     env,
			"Secrets":         secrets,
			"ServiceAccount":  fn.ServiceAccount,
			"Resources":       resources,
			"MinReplicas":     minReplicas,
			"MaxReplicas":     maxReplicas,
		}
		out, err := m.executeFunction("function", fn.Key, data)
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

	for name, r := range routes {
		if len(r) == 0 {
			continue
		}

		fqdn := r[0].Route.FQDN
		upstreams := make(map[string]bool)

		for _, inner := range r {
			if inner.Route.FQDN != fqdn {
				return nil, ErrInvalidFQDN
			}

			upstreams[inner.Key] = true
		}

		data := map[string]interface{}{
			"Key":       "routes--" + name,
			"Namespace": m.namespace,
			"Commit":    m.commit,
			"FQDN":      fqdn,
			"Upstreams": upstreams,
			"Routes":    r,
		}
		out, err := m.executeFunction("proxy", name, data)
		if err != nil {
			return nil, err
		}
		all = append(all, out...)

	}

	return all, nil
}

func NewManager(workingDir string, namespace string, commit string) Manager {
	namespacePrefix := ""
	routesPrefix := ""
	indexOf := strings.Index(namespace, "--")
	if indexOf > -1 {
		namespacePrefix = namespace[0 : indexOf+2]
		routesPrefix = namespace[indexOf+2:] + "--"
	}

	funcMap := GetFuncMaps(namespacePrefix, routesPrefix)

	return &manager{
		workingDir: workingDir,
		namespace:  namespace,
		commit:     commit,
		funcMap:    funcMap,
	}
}
