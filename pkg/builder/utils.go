package builder

import (
	"archive/tar"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/contextcloud/ccb/pkg/print"
)

var ignoreNames = []string{".gitignore", ".dockerignore"}

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

func buildArchive(infos ...*ArchiveInfo) (io.Reader, error) {
	// Create a buffer
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	for _, info := range infos {
		if err := info.Write(tw); err != nil {
			return nil, err
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
