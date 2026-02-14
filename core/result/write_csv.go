package result

import (
	"bgscan/core/filemanager"
)

func WriteCSV(
	path string,
	fn func(func(ResultIPScan) error) error,
) error {
	return filemanager.StreamWriteCSV(path, csvCfg, func(write func([]string) error) error {
		return fn(func(r ResultIPScan) error {
			return write(r.EncodeCSV())
		})
	})
}
