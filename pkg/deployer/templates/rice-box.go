// Code generated by rice embed-go; DO NOT EDIT.
package templates

import (
	"time"

	"github.com/GeertJohan/go.rice/embedded"
)

func init() {

	// define files
	file3 := &embedded.EmbeddedFile{
		Filename:    "function/deployment.yaml",
		FileModTime: time.Unix(1648536937, 0),

		Content: string("apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: {{ .Key }}\n  namespace: {{ .Namespace }}\n  labels: \n    release: {{ .Name }}\n    version: {{ .Version | quote }}\n    commit: {{ .Commit | quote }}\nspec:\n{{- if $.Replicas }}\n  replicas: {{ .Replicas }}\n{{- end }}\n  revisionHistoryLimit: 10\n  selector:\n    matchLabels:\n      release: {{ .Name }}\n      version: {{ .Version | quote }}\n  template:\n    metadata:\n      annotations:\n        linkerd.io/inject: enabled\n      name: {{ .Key }}\n      labels:\n        release: {{ .Name }}\n        version: {{ .Version | quote }}\n        commit: {{ .Commit | quote }}\n        {{- with .Labels }}\n        {{- toYaml . | nindent 8 }}\n        {{- end }}\n    spec:\n      containers:\n      - name: {{ .Name }}\n        image: {{ .Image }}\n{{- if $.Secrets }}\n        envFrom:\n        {{- range $key, $value := .Secrets }}\n        - secretRef:\n            name: {{ $value }}\n        {{- end }}\n{{- end }}\n{{- if $.Environment }}\n        env:\n        {{- range $key, $value := .Environment }}\n        - name: {{ $key }}\n          value: {{ $value | quote }}\n        {{- end }}\n{{- end }}\n        ports:\n        - name: http\n          containerPort: 8080\n        - name: metrics\n          containerPort: 8081\n        - name: health\n          containerPort: 8082\n{{- if $.LivenessProbe.Enabled }}\n        livenessProbe:\n          httpGet:\n            path: {{ $.LivenessProbe.Path }}\n            port: {{ $.LivenessProbe.Port }}\n            scheme: HTTP\n          initialDelaySeconds: {{ $.LivenessProbe.InitialDelaySeconds }}\n          timeoutSeconds: {{ $.LivenessProbe.TimeoutSeconds }}\n          periodSeconds: {{ $.LivenessProbe.PeriodSeconds }}\n{{- end }}\n{{- if $.ReadinessProbe.Enabled }}\n        readinessProbe:\n          httpGet:\n            path: {{ $.ReadinessProbe.Path }}\n            port: {{ $.ReadinessProbe.Port }}\n            scheme: HTTP\n          initialDelaySeconds: {{ $.ReadinessProbe.InitialDelaySeconds }}\n          timeoutSeconds: {{ $.ReadinessProbe.TimeoutSeconds }}\n          periodSeconds: {{ $.ReadinessProbe.PeriodSeconds }}\n{{- end }}\n{{- if or $.Limits $.Requests }}\n        resources:\n{{- if $.Limits }}\n          limits:\n            {{- range $key, $value := .Limits }}\n            {{ $key }}: {{ $value }}\n            {{- end }}\n{{- end }}\n{{- if $.Requests }}\n          requests:\n            {{- range $key, $value := .Requests }}\n            {{ $key }}: {{ $value }}\n            {{- end }}\n{{- end }}\n{{- end }}\n{{- if $.ReadOnlyRootFilesystem }}\n        securityContext:\n          readOnlyRootFilesystem: {{ $.ReadOnlyRootFilesystem }}\n{{- end }}\n        volumeMounts:\n        - mountPath: /tmp\n          name: temp\n{{- if $.Secrets }}\n      {{- range $key, $value := .Secrets }}\n        - mountPath: /var/secrets\n          name: {{ $value }}\n          readOnly: true\n      {{- end }}\n{{- end }}\n{{- if $.NodeSelector }}\n      nodeSelector:\n        {{- range $key, $value := .NodeSelector }}\n        {{ $key }}: {{ $value }}\n        {{- end }}\n{{- end }}\n      volumes:\n      - emptyDir: {}\n        name: temp\n{{- if $.Secrets }}\n    {{- range $key, $value := .Secrets }}\n      - name: {{ $value }}\n        projected:\n          defaultMode: 420\n          sources:\n          - secret:\n              name: {{ $value }}\n    {{- end }}\n{{- end }}"),
	}
	file4 := &embedded.EmbeddedFile{
		Filename:    "function/service.yaml",
		FileModTime: time.Unix(1648455884, 0),

		Content: string("apiVersion: v1\nkind: Service\nmetadata:\n  name: {{ .Key }}\n  namespace: {{ .Namespace }}\n  labels: \n    release: {{ .Name }}\n    version: {{ .Version | quote }}\n    commit: {{ .Commit | quote }}\n{{- if $.Annotations }}\n  annotations:\n    {{- toYaml .Annotations | nindent 4 }}        \n{{- end }}\nspec:\n  ports:\n    - name: http\n      port: 8080\n    - name: metrics\n      port: 8081\n    - name: health\n      port: 8082\n  selector:\n    release: {{ .Name }}\n    version: {{ .Version | quote }}"),
	}
	file6 := &embedded.EmbeddedFile{
		Filename:    "includes/certificate.yaml",
		FileModTime: time.Unix(1648455884, 0),

		Content: string("apiVersion: cert-manager.io/v1\nkind: Certificate\nmetadata:\n  name: {{ .Key }}\n  namespace: {{ .Namespace }}\n  labels: \n    commit: {{ .Commit | quote }}\nspec:\n  dnsNames:\n  - {{ .FQDN | quote }}\n  issuerRef:\n    group: cert-manager.io\n    kind: ClusterIssuer\n    name: letsencrypt\n  secretName: {{ .Key }}"),
	}
	file7 := &embedded.EmbeddedFile{
		Filename:    "includes/includes.yaml",
		FileModTime: time.Unix(1648872813, 0),

		Content: string("apiVersion: projectcontour.io/v1\nkind: HTTPProxy\nmetadata:\n  name: {{ .Key }}\n  namespace: {{ .Namespace }}\n  labels: \n    commit: {{ .Commit | quote }}\nspec:\n  virtualhost:\n    fqdn: {{ .FQDN | quote }}\n    tls:\n      secretName: {{ .Key }}\n{{- if $.Routes }}\n  routes:\n  {{- range $key, $value := .Routes }}\n    - conditions:\n      - prefix: {{ $value.Prefix }}\n    {{- if $value.Headers }}\n      {{- range $key, $value := $value.Headers }}\n      - header:\n          name: {{ $value.Name }}\n      {{- end }}\n    {{- end }}\n    {{- if $value.Redirect }}\n      requestRedirectPolicy:\n        hostname: {{ $value.Redirect }}\n    {{- else }}\n      services:\n        - name: {{ $value.Key }}\n          port: 8080\n    {{- end }}    \n  {{- end }}\n{{- end }}\n{{- if .Includes }}\n  includes:\n  {{- range $key, $value := .Includes }}\n    - conditions:\n      - prefix: {{ $value.Prefix }}\n    {{- if $value.Headers }}\n      {{- range $key, $value := $value.Headers }}\n      - header:\n          name: {{ $value.Name }}\n      {{- end }}\n    {{- end }}\n      name: {{ $value.Name | route }}\n    {{- if $.Namespace }}\n      namespace: {{ .Namespace | namespace }}\n    {{- end }}\n  {{- end }}\n{{- end }}"),
	}
	file9 := &embedded.EmbeddedFile{
		Filename:    "proxy/proxy.yaml",
		FileModTime: time.Unix(1648872964, 0),

		Content: string("\napiVersion: projectcontour.io/v1\nkind: HTTPProxy\nmetadata:\n  name: {{ .Key }}\n  namespace: {{ .Namespace }}\n  labels: \n    commit: {{ .Commit | quote }}\nspec:\n  routes:\n  {{- range $key, $value := .Routes }}\n    - conditions:\n      - prefix: {{ $value.Route.Prefix }}\n    {{- if $value.Route.Headers }}\n      {{- range $key, $value := $value.Route.Headers }}\n      - header:\n          name: {{ $value.Name }}\n      {{- end }}\n    {{- end }}\n    {{- if $value.Route.Redirect }}\n      requestRedirectPolicy:\n        hostname: {{ $value.Route.Redirect }}\n    {{- else }}\n      services:\n        - name: {{ $value.Key }}\n          port: 8080\n    {{- end }}\n  {{- end }}"),
	}
	filea := &embedded.EmbeddedFile{
		Filename:    "templates.go",
		FileModTime: time.Unix(1648455884, 0),

		Content: string("//go:generate rice embed-go\n\npackage templates\n\nimport (\n\t\"strings\"\n\t\"text/template\"\n\n\trice \"github.com/GeertJohan/go.rice\"\n\t\"github.com/Masterminds/sprig\"\n\t\"gopkg.in/yaml.v2\"\n)\n\nfunc NewBox() *rice.Box {\n\tconf := rice.Config{\n\t\tLocateOrder: []rice.LocateMethod{rice.LocateEmbedded, rice.LocateAppended, rice.LocateFS},\n\t}\n\treturn conf.MustFindBox(\".\")\n}\n\nfunc toYAML(v interface{}) string {\n\tdata, err := yaml.Marshal(v)\n\tif err != nil {\n\t\t// Swallow errors inside of a template.\n\t\treturn \"\"\n\t}\n\treturn strings.TrimSuffix(string(data), \"\\n\")\n}\n\nfunc GetFuncMaps(namespacePrefix string, routePrefix string) template.FuncMap {\n\tfm := sprig.TxtFuncMap()\n\tfm[\"toYaml\"] = toYAML\n\tfm[\"namespace\"] = func(v interface{}) string {\n\t\tns, ok := v.(string)\n\t\tif !ok {\n\t\t\treturn \"\"\n\t\t}\n\t\treturn namespacePrefix + ns\n\t}\n\tfm[\"route\"] = func(v interface{}) string {\n\t\tns, ok := v.(string)\n\t\tif !ok {\n\t\t\treturn \"\"\n\t\t}\n\t\treturn routePrefix + ns\n\t}\n\treturn fm\n}\n"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   "",
		DirModTime: time.Unix(1648455884, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			filea, // "templates.go"

		},
	}
	dir2 := &embedded.EmbeddedDir{
		Filename:   "function",
		DirModTime: time.Unix(1648455884, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file3, // "function/deployment.yaml"
			file4, // "function/service.yaml"

		},
	}
	dir5 := &embedded.EmbeddedDir{
		Filename:   "includes",
		DirModTime: time.Unix(1648455884, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file6, // "includes/certificate.yaml"
			file7, // "includes/includes.yaml"

		},
	}
	dir8 := &embedded.EmbeddedDir{
		Filename:   "proxy",
		DirModTime: time.Unix(1648455884, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file9, // "proxy/proxy.yaml"

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{
		dir2, // "function"
		dir5, // "includes"
		dir8, // "proxy"

	}
	dir2.ChildDirs = []*embedded.EmbeddedDir{}
	dir5.ChildDirs = []*embedded.EmbeddedDir{}
	dir8.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`.`, &embedded.EmbeddedBox{
		Name: `.`,
		Time: time.Unix(1648455884, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"":         dir1,
			"function": dir2,
			"includes": dir5,
			"proxy":    dir8,
		},
		Files: map[string]*embedded.EmbeddedFile{
			"function/deployment.yaml":  file3,
			"function/service.yaml":     file4,
			"includes/certificate.yaml": file6,
			"includes/includes.yaml":    file7,
			"proxy/proxy.yaml":          file9,
			"templates.go":              filea,
		},
	})
}
