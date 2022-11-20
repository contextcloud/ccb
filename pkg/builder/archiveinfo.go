package builder

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/denormal/go-gitignore"
)

type ArchiveInfo struct {
	Type   string
	Name   string
	Folder string
	Body   []byte
}

func (info *ArchiveInfo) addFile(tw *tar.Writer) error {
	// Make a TAR header for the file
	tarHeader := &tar.Header{
		Name: info.Name,
		Size: int64(len(info.Body)),
	}
	if err := tw.WriteHeader(tarHeader); err != nil {
		return err
	}
	// Writes the dockerfile data to the TAR file
	if _, err := tw.Write(info.Body); err != nil {
		return err
	}
	return nil
}

func (info *ArchiveInfo) addDir(tw *tar.Writer) error {
	// Make a TAR header for the file
	var ignores []gitignore.GitIgnore
	for _, filename := range ignoreNames {
		ignorePath := path.Join(info.Folder, filename)
		ignore, err := gitignore.NewFromFile(ignorePath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if ignore != nil {
			ignores = append(ignores, ignore)
		}
	}

	basename := path.Base(info.Folder)

	walkFn := func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.Mode().IsDir() {
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
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			if link, err = os.Readlink(p); err != nil {
				return err
			}
		}

		h, err := tar.FileInfoHeader(fi, link)
		if err != nil {
			return err
		}

		n := strings.TrimPrefix(p, info.Folder)
		h.Name = path.Join(basename, filepath.ToSlash(n))

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

	if err := filepath.Walk(info.Folder, walkFn); err != nil {
		return err
	}

	return nil
}

func (info *ArchiveInfo) Write(tw *tar.Writer) error {
	switch info.Type {
	case "dir":
		return info.addDir(tw)
	case "raw":
		return info.addFile(tw)
	default:
		return fmt.Errorf("Unknown archive type: %s", info.Type)
	}
}

func NewDirArchive(folder string) *ArchiveInfo {
	name := path.Base(folder)
	return &ArchiveInfo{
		Type:   "dir",
		Name:   name,
		Folder: folder,
	}
}

func NewRawArchive(name string, body []byte) *ArchiveInfo {
	return &ArchiveInfo{
		Type: "raw",
		Name: name,
		Body: body,
	}
}
