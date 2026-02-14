package filemanager

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ═══════════════════════════════════════════════════════════
// Text File Operations
// ═══════════════════════════════════════════════════════════

// WriteTextFile writes plain text to a file (creates dirs if missing)
func WriteTextFile(path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// WriteTextFileIfNotExist writes text only if file doesn't exist
func WriteTextFileIfNotExist(path string, content string) error {
	if CheckFileExists(path) {
		return nil
	}
	return WriteTextFile(path, content)
}

// GetTextFile reads a text file and returns its content
func GetTextFile(path string) (string, error) {
	if !CheckFileExists(path) {
		return "", fmt.Errorf("file does not exist: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return string(data), nil
}

// AppendTextFile appends text to a file (creates dirs if missing)
func AppendTextFile(path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("failed to append to file %s: %w", path, err)
	}

	return nil
}

// ═══════════════════════════════════════════════════════════
// Streaming Text Files
// ═══════════════════════════════════════════════════════════

// TextStreamConfig controls scanner behavior while streaming.
type TextStreamConfig struct {
	SplitFunc  bufio.SplitFunc // defaults to bufio.ScanLines
	BufferSize int             // default 64 KiB; increase for very long lines
	MaxToken   int             // max token size; defaults to BufferSize
}

// StreamTextFile reads the file line-by-line (or token-by-token) and calls handler for each piece.
// Handler returning an error stops the stream and propagates the error.
func StreamTextFile(path string, cfg TextStreamConfig, handler func(string) error) error {
	if !CheckFileExists(path) {
		return fmt.Errorf("file does not exist: %s", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open text file %s: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	applyTextStreamConfig(scanner, cfg)

	for scanner.Scan() {
		if err := handler(scanner.Text()); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("text stream error: %w", err)
	}

	return nil
}

// StreamTextToChan pushes each scanned token (default: line) into the provided channel.
// The caller is responsible for closing the channel after StreamTextToChan returns.
func StreamTextToChan(path string, cfg TextStreamConfig, out chan<- string) error {
	return StreamTextFile(path, cfg, func(token string) error {
		out <- token
		return nil
	})
}

func applyTextStreamConfig(scanner *bufio.Scanner, cfg TextStreamConfig) {
	if cfg.SplitFunc != nil {
		scanner.Split(cfg.SplitFunc)
	} else {
		scanner.Split(bufio.ScanLines)
	}

	bufSize := cfg.BufferSize
	if bufSize <= 0 {
		bufSize = 64 * 1024 // 64 KiB default buffer
	}

	maxToken := cfg.MaxToken
	if maxToken <= 0 {
		maxToken = bufSize
	}

	buffer := make([]byte, bufSize)
	scanner.Buffer(buffer, maxToken)
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(dst)
		return err
	}
	if err := out.Sync(); err != nil {
		_ = out.Close()
		_ = os.Remove(dst)
		return err
	}
	return out.Close()
}
