package result

import (
	"bgscan/internal/core/filemanager"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Result directories used by different scan types.
//
// Each directory stores CSV files produced by a specific scan engine.
const (
	ICMPResultDir       = "result/icmp/"
	TCPResultDir        = "result/tcp/"
	HTTPResultDir       = "result/http/"
	XRAYResultDir       = "result/xray/"
	ResolveResultDir    = "result/resolve/"
	DNSTTResultDir      = "result/dnstt/"
	SlipStreamResultDir = "result/slipstream/"
)

// ListResultFiles returns metadata for all result CSV files matching
// the given scan type.
//
// For performance reasons, this function only reads filesystem metadata
// and does NOT count the number of IPs stored inside each file.
// As a result, the IPCount field is always set to -1.
//
// Use CountIPsInFile if an accurate IP count is required.
func ListResultFiles(searchType ResultType) ([]ResultFile, error) {
	dirs := resolveResultDirs(searchType)
	if len(dirs) == 0 {
		return nil, nil
	}

	var results []ResultFile

	for _, d := range dirs {
		entries, err := os.ReadDir(d.dir)
		if err != nil {
			// Directory may not exist or be inaccessible.
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			if !strings.HasSuffix(strings.ToLower(name), ".csv") {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}

			results = append(results, ResultFile{
				Name:        filemanager.StripExt(name),
				SizeBytes:   info.Size(),
				CreatedTime: info.ModTime(),
				Type:        d.rType,
				Path:        filepath.Join(d.dir, name),
				IPCount:     -1, // intentionally skipped
			})
		}
	}

	return results, nil
}

// GetResultFileInfo returns metadata for a single result file path.
//
// The returned ResultFile contains filesystem metadata only.
// IPCount is set to -1 and must be calculated explicitly if needed.
func GetResultFileInfo(path string) (ResultFile, error) {
	info, err := os.Stat(path)
	if err != nil {
		return ResultFile{}, fmt.Errorf("cannot read result file: %w", err)
	}

	return ResultFile{
		Name:        filemanager.StripExt(info.Name()),
		SizeBytes:   info.Size(),
		CreatedTime: info.ModTime(),
		Type:        ResultTypeFromPath(path),
		Path:        path,
		IPCount:     -1, // intentionally skipped
	}, nil
}

// NormalizeResultFileName ensures that a result file name has the `.csv` extension.
//
// If the provided name does not already end with `.csv`, the extension
// will be appended automatically. Existing `.csv` extensions are preserved.
//
// Example:
//
//	NormalizeResultFileName("scan_1")     → "scan_1.csv"
//	NormalizeResultFileName("scan_1.csv") → "scan_1.csv"
func NormalizeResultFileName(name string) string {
	if !filemanager.HasExt(name, ".csv") {
		return name + ".csv"
	}
	return name
}

// ResultTypeFromPath infers the result type from a filesystem path.
//
// The detection is based on known result directory segments.
// If no match is found, ResultICMP is used as a safe default.
func ResultTypeFromPath(path string) ResultType {
	switch {
	case strings.Contains(path, ICMPResultDir):
		return ResultICMP
	case strings.Contains(path, TCPResultDir):
		return ResultTCP
	case strings.Contains(path, HTTPResultDir):
		return ResultHTTP
	case strings.Contains(path, ResolveResultDir):
		return ResultRESOLVE
	case strings.Contains(path, DNSTTResultDir):
		return ResultDNSTT
	case strings.Contains(path, SlipStreamResultDir):
		return ResultSLIPSTREAM
	case strings.Contains(path, XRAYResultDir):
		return ResultXRAY
	default:
		return ResultICMP
	}
}

// CountIPsInFile counts the number of IP entries inside a result file.
//
// This operation reads the file contents and may be relatively expensive
// for large result sets.
func CountIPsInFile(file ResultFile) (int64, error) {
	path := filepath.Join(resolveDir(file.Type), file.Name)
	return Count(path)
}

// -----------------------------------------------------------------------------
// Internal helpers
// -----------------------------------------------------------------------------

type resultDir struct {
	dir   string
	rType ResultType
}

// resolveResultDirs returns filesystem directories matching the
// requested result type.
func resolveResultDirs(searchType ResultType) []resultDir {
	all := []resultDir{
		{ICMPResultDir, ResultICMP},
		{TCPResultDir, ResultTCP},
		{HTTPResultDir, ResultHTTP},
		{XRAYResultDir, ResultXRAY},
		{DNSTTResultDir, ResultDNSTT},
		{SlipStreamResultDir, ResultSLIPSTREAM},
		{ResolveResultDir, ResultRESOLVE},
	}

	if searchType == ResultAll {
		return all
	}

	for _, d := range all {
		if d.rType == searchType {
			return []resultDir{d}
		}
	}

	return nil
}

// resolveDir returns the directory path associated with a result type.
func resolveDir(rType ResultType) string {
	switch rType {
	case ResultICMP:
		return ICMPResultDir
	case ResultTCP:
		return TCPResultDir
	case ResultHTTP:
		return HTTPResultDir
	case ResultXRAY:
		return XRAYResultDir
	case ResultRESOLVE:
		return ResolveResultDir
	case ResultDNSTT:
		return DNSTTResultDir
	case ResultSLIPSTREAM:
		return SlipStreamResultDir
	default:
		return ""
	}
}

// BuildResultFilePath generates a new result file path using a prefix
// and the current timestamp.
//
// Example:
//
//	result/tcp/tcp_20260315_143022.csv
//
// The prefix typically identifies the scan target or configuration.
func BuildResultFilePath(resultDir, prefix string) (string, error) {
	switch resultDir {
	case ICMPResultDir, TCPResultDir, HTTPResultDir, XRAYResultDir, DNSTTResultDir, ResolveResultDir, SlipStreamResultDir:
		// valid directory
	default:
		return "", fmt.Errorf("invalid result directory")
	}

	if prefix == "" {
		return "", fmt.Errorf("prefix cannot be empty")
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s%s.csv", prefix, timestamp)

	return filepath.Join(resultDir, filename), nil
}
