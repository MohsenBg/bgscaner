package ipmanager

import (
	"os"
	"sort"
	"time"

	"bgscan/core/filemanager"
	"bgscan/core/iplist"
)

type IPFileInfo struct {
	Name      string // without .csv
	Path      string // absolute path
	Size      int64  // bytes
	CreatedAt time.Time
}

const IPListDir = "ips"

var ipCSVConfig = filemanager.CSVConfig{
	Comma:           ',',
	HasHeader:       false,
	FieldsPerRecord: -1,
	LazyQuotes:      true,
}

func AddIPFile(srcPath, dstPath string) error {
	return iplist.WriteCSV(dstPath, func(write func(iplist.IPList) error) error {
		return iplist.ReadCSV(srcPath, func(ip iplist.IPList) error {
			return write(ip)
		})
	})
}

func ListIPFiles() ([]IPFileInfo, error) {
	files, err := filemanager.ListFiles(
		IPListDir,
		func(name string, info os.FileInfo) bool {
			return filemanager.HasExt(name, ".csv")
		},
	)
	if err != nil {
		return nil, err
	}

	out := make([]IPFileInfo, 0, len(files))

	for _, f := range files {
		out = append(out, IPFileInfo{
			Name:      filemanager.StripExt(f.Name),
			Path:      f.Path,
			Size:      f.Info.Size(),
			CreatedAt: f.Info.ModTime(),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})

	return out, nil
}
