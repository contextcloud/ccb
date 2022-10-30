package deployer

import (
	"os"

	"github.com/contextcloud/ccb/pkg/print"
)

type ManifestType string

var (
	DeploymentManifestType  ManifestType = "Deployment"
	ServiceManifestType     ManifestType = "Service"
	SecretManifestType      ManifestType = "Secret"
	ProxyManifestType       ManifestType = "Proxy"
	CertificateManifestType ManifestType = "Certificate"
	VirtualServerType       ManifestType = "VirtualServer"
)

func ToManifestType(p string) ManifestType {
	switch p {
	case "function/deployment.yaml":
		return DeploymentManifestType
	case "function/service.yaml":
		return ServiceManifestType
	case "proxy/proxy.yaml":
		return ProxyManifestType
	case "routes/certificate.yaml":
		return CertificateManifestType
	case "routes/server.yaml":
		return VirtualServerType
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

func (m Manifests) Print(l print.Log) {
	all := m.merged()
	l.Print(all)
}
