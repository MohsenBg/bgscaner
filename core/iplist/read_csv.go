package iplist

import (
	"bgscan/core/filemanager"
	"bgscan/core/ip"
	"strings"
)

// shared CSV configuration for ip files
var DefaultCSVConfig = filemanager.CSVConfig{
	Comma:           ',',
	HasHeader:       false,
	FieldsPerRecord: -1,
	LazyQuotes:      true,
}

// ReadCSV streams IPList records from CSV file.
func ReadCSV(
	path string,
	fn func(IPList) error,
) error {
	return filemanager.StreamCSV(path, DefaultCSVConfig, func(rec []string) error {
		ip, ok := ParseCSV(rec)
		if !ok {
			return nil // skip invalid row
		}
		return fn(ip)
	})
}

func StreamActiveIPs(path string, cfg filemanager.CSVConfig, out chan<- string) error {
	return filemanager.StreamCSV(path, cfg, func(rec []string) error {
		row, ok := ParseCSV(rec)
		if !ok || !row.Enable {
			return nil
		}

		if strings.Contains(row.IP, "/") {
			return ip.StreamCIDR(row.IP, out)
		}

		out <- row.IP
		return nil
	})
}
