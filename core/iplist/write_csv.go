package iplist

import (
	"bgscan/core/filemanager"
)

func WriteCSV(
	path string,
	fn func(func(IPList) error) error,
) error {
	return filemanager.StreamWriteCSV(path, DefaultCSVConfig, func(write func([]string) error) error {
		return fn(func(ip IPList) error {
			return write(ip.EncodeCSV())
		})
	})
}
