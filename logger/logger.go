package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sanity-io/litter"
)

var (
	logFile *os.File
	enabled = true
)

// Init
func Init(filename string) error {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// timestamp := time.Now().Format("2006-01-02_15-04-05")
	logPath := filepath.Join(logDir, "app.log")

	var err error
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	Log("=== Debug session started ===")
	return nil
}

// Close
func Close() {
	if logFile != nil {
		Log("=== Debug session ended ===")
		logFile.Close()
	}
}

// Log
func Log(msg string, args ...interface{}) {
	if !enabled || logFile == nil {
		return
	}

	timestamp := time.Now().Format("15:04:05.000")
	line := fmt.Sprintf("[%s] %s\n", timestamp, fmt.Sprintf(msg, args...))
	logFile.WriteString(line)
}

// Dump
func Dump(label string, v interface{}) {
	if !enabled || logFile == nil {
		return
	}

	timestamp := time.Now().Format("15:04:05.000")
	fmt.Fprintf(logFile, "\n[%s] === %s ===\n", timestamp, label)

	litter.Config.HidePrivateFields = false
	litter.Config.FieldExclusions = regexp.MustCompile(`^XXX_`)
	output := litter.Sdump(v)

	logFile.WriteString(output + "\n")
}

// DumpShort
func DumpShort(label string, v interface{}) {
	if !enabled || logFile == nil {
		return
	}

	timestamp := time.Now().Format("15:04:05.000")
	line := fmt.Sprintf("[%s] %s: %#v\n", timestamp, label, v)
	logFile.WriteString(line)
}

// Enable/Disable
func Enable()         { enabled = true }
func Disable()        { enabled = false }
func IsEnabled() bool { return enabled }

// Separator
func Separator() {
	if !enabled || logFile == nil {
		return
	}
	logFile.WriteString("\n" + strings.Repeat("─", 80) + "\n\n")
}
