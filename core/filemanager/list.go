package filemanager

import (
	"os"
	"path/filepath"
)

type FileEntry struct {
	Name string
	Path string
	Info os.FileInfo
}

func ListFiles(
	dir string,
	filter func(name string, info os.FileInfo) bool,
) ([]FileEntry, error) {

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]FileEntry, 0, len(entries))

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		info, err := e.Info()
		if err != nil {
			continue
		}

		if filter != nil && !filter(e.Name(), info) {
			continue
		}

		abs, err := filepath.Abs(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}

		files = append(files, FileEntry{
			Name: e.Name(),
			Path: abs,
			Info: info,
		})
	}

	return files, nil
}
