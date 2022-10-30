package builder

import (
	"archive/tar"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/contextcloud/ccb/pkg/print"
	"github.com/denormal/go-gitignore"
)

var ignoreNames = []string{".gitignore", ".dockerignore"}

func pstring(s string) *string {
	return &s
}

type BuildLine struct {
	Stream      string            `json:"stream"`
	Aux         *BuildAux         `json:"aux"`
	Error       string            `json:"error"`
	ErrorDetail *BuildErrorDetail `json:"errorDetail"`
}
type BuildErrorDetail struct {
	Message string `json:"message"`
}
type BuildAux struct {
	Id string `json:"id"`
}

type PushLine struct {
	Status string   `json:"status"`
	Aux    *PushAux `json:"aux"`
	Error  string   `json:"error"`
}
type PushAux struct {
	Tag    string `json:"tag"`
	Digest string `json:"digest"`
	Size   int64  `json:"size"`
}

type ArchiveInfo struct {
	Type     string
	Name     string
	Relative bool
	Body     []byte
}

func NewDirArchive(name string, relative bool) *ArchiveInfo {
	return &ArchiveInfo{
		Type:     "dir",
		Name:     name,
		Relative: relative,
	}
}

func NewRawArchive(name string, body []byte) *ArchiveInfo {
	return &ArchiveInfo{
		Type: "raw",
		Name: name,
		Body: body,
	}
}

func addFile(tw *tar.Writer, name string, body []byte) error {
	// Make a TAR header for the file
	tarHeader := &tar.Header{
		Name: name,
		Size: int64(len(body)),
	}
	if err := tw.WriteHeader(tarHeader); err != nil {
		return err
	}
	// Writes the dockerfile data to the TAR file
	if _, err := tw.Write(body); err != nil {
		return err
	}
	return nil
}

func addDir(tw *tar.Writer, folder string, relative bool) error {
	// Make a TAR header for the file
	var ignores []gitignore.GitIgnore
	for _, filename := range ignoreNames {
		ignorePath := path.Join(folder, filename)
		ignore, err := gitignore.NewFromFile(ignorePath)
		if err != nil && !os.IsNotExist(err) {
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

		// skip stuff
		for _, ign := range ignores {
			m := ign.Match(p)
			if m != nil && m.Ignore() {
				return nil
			}
		}

		var link string
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			if link, err = os.Readlink(p); err != nil {
				return err
			}
		}

		name := p
		if relative {
			name = strings.TrimPrefix(name, folder)
		}
		name = filepath.ToSlash(name)

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

	return nil
}

func buildArchive(infos ...*ArchiveInfo) (io.Reader, error) {
	// Create a buffer
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	for _, info := range infos {
		switch info.Type {
		case "dir":
			if err := addDir(tw, info.Name, info.Relative); err != nil {
				return nil, err
			}
			break
		case "raw":
			if err := addFile(tw, info.Name, info.Body); err != nil {
				return nil, err
			}
			break
		default:
			return nil, fmt.Errorf("Unknown archive type: %s", info.Type)
		}
	}

	// return the reader
	reader := bytes.NewReader(buf.Bytes())
	return reader, nil
}

func buildResult(rd io.Reader, info print.Log) ([]*BuildAux, error) {
	var out []*BuildAux

	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		line := scanner.Bytes()

		item := &BuildLine{}
		if err := json.Unmarshal(line, item); err != nil {
			return nil, err
		}

		// parse stuff
		if item.Error != "" {
			return nil, errors.New(item.Error)
		}
		if item.Stream != "" {
			info.Print(item.Stream)
		}
		if item.Aux != nil {
			out = append(out, item.Aux)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func pushResult(rd io.Reader, info print.Log) ([]*PushAux, error) {
	var out []*PushAux

	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		line := scanner.Bytes()

		item := &PushLine{}
		if err := json.Unmarshal(line, item); err != nil {
			return nil, err
		}

		// parse stuff
		if item.Error != "" {
			return nil, errors.New(item.Error)
		}
		if item.Status != "" {
			info.Println(item.Status)
		}
		if item.Aux != nil {
			out = append(out, item.Aux)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
