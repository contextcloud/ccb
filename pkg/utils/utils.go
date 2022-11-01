package utils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func ParseMap(envvars []string, keyName string) (map[string]*string, error) {
	result := make(map[string]*string)
	for _, envvar := range envvars {
		s := strings.SplitN(strings.TrimSpace(envvar), "=", 2)
		if len(s) != 2 {
			return nil, fmt.Errorf("label format is not correct, needs key=value")
		}
		envvarName := s[0]
		envvarValue := s[1]

		if !(len(envvarName) > 0) {
			return nil, fmt.Errorf("empty %s name: [%s]", keyName, envvar)
		}
		if !(len(envvarValue) > 0) {
			return nil, fmt.Errorf("empty %s value: [%s]", keyName, envvar)
		}

		result[envvarName] = &envvarValue
	}
	return result, nil
}

func MergeMap[T any](i map[string]T, j map[string]T) map[string]T {
	merged := make(map[string]T)

	for k, v := range i {
		merged[k] = v
	}
	for k, v := range j {
		merged[k] = v
	}
	return merged
}

func YamlFile(workingDir string, filename string) (string, error) {
	name := path.Join(workingDir, filename)
	ext := filepath.Ext(filename)

	name = name[0 : len(name)-len(ext)]
	name = path.Clean(name)

	paths := []string{
		name + ".yaml",
		name + ".yml",
	}

	for _, p := range paths {
		info, err := os.Stat(p)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}

		if info.IsDir() {
			return "", fmt.Errorf("Is a directory: %s", p)
		}

		return p, nil
	}

	return "", os.ErrNotExist
}
