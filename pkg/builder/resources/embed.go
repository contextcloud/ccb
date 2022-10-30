package resources

import (
	_ "embed"
)

//go:embed files.Dockerfile
var FilesDockerFile []byte
