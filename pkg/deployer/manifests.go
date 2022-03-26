package deployer

import (
	"os"

	"github.com/contextcloud/ccb-cli/pkg/print"
)

type ManifestType string

var (
	DeploymentManifestType ManifestType = "Deployment"
	ServiceManifestType    ManifestType = "Service"
	SecretManifestType     ManifestType = "Secret"
	ProxyManifestType      ManifestType = "Proxy"
)

func ToManifestType(p string) ManifestType {
	switch p {
	case "function/deployment.yaml":
		return DeploymentManifestType
	case "function/service.yaml":
		return ServiceManifestType
	case "proxy/proxy.yaml":
		return ProxyManifestType
	default:
		return ""
	}
}

type Manifest struct {
	Type    ManifestType
	Key     string
	Content string
}

type Manifests []Manifest

func (m Manifests) merged() string {
	var all string
	for i, manifest := range m {
		if i > 0 {
			all += "\n"
		}
		all += "---\n" + manifest.Content
	}
	return all
}

func (m Manifests) Save(filename string) error {
	all := m.merged()
	return os.WriteFile(filename, []byte(all), 0644)
}

func (m Manifests) Print(out print.Out) {
	all := m.merged()
	out.Print(all)
}
