package result

import (
	"bgscan/core/filemanager"
)

var csvCfg = filemanager.CSVConfig{Comma: ','}

func ReadCSV(path string, fn func(ResultIPScan) error) error {
	return filemanager.StreamCSV(path, csvCfg, func(rec []string) error {
		r, ok := ParseCSV(rec)
		if !ok {
			return nil
		}
		return fn(r)
	})
}
