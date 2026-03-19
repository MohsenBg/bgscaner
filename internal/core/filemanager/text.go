package filemanager

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// ═══════════════════════════════════════════════════════════
// Text File Operations
// ═══════════════════════════════════════════════════════════

// WriteTextFile writes plain text to a file.
// The directory will be created if it does not exist.
func WriteTextFile(path string, content string) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write text file %s: %w", path, err)
	}

	return nil
}

// WriteTextFileIfNotExist writes text only if the file does not already exist.
func WriteTextFileIfNotExist(path string, content string) error {
	if CheckFileExists(path) {
		return nil
	}
	return WriteTextFile(path, content)
}

// GetTextFile reads a text file and returns its content.
func GetTextFile(path string) (string, error) {
	if !CheckFileExists(path) {
		return "", fmt.Errorf("text file does not exist: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read text file %s: %w", path, err)
	}

	return string(data), nil
}

// AppendTextFile appends text to a file.
// The directory will be created if it does not exist.
func AppendTextFile(path string, content string) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open text file %s: %w", path, err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("append text file %s: %w", path, err)
	}

	return nil
}

// ═══════════════════════════════════════════════════════════
// Streaming Text Files
// ═══════════════════════════════════════════════════════════

// TextStreamConfig controls scanner behavior while streaming text.
type TextStreamConfig struct {
	SplitFunc  bufio.SplitFunc // defaults to bufio.ScanLines
	BufferSize int             // initial buffer size (default: 64 KiB)
	MaxToken   int             // maximum token size (default: BufferSize)
}

// StreamTextFile reads the file token-by-token (default: line-by-line)
// and calls handler for each token.
// Returning an error from handler stops the stream.
func StreamTextFile(
	path string,
	cfg TextStreamConfig,
	handler func(string) error,
) error {

	if !CheckFileExists(path) {
		return fmt.Errorf("text file does not exist: %s", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open text file %s: %w", path, err)
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

// StreamTextToChan sends each scanned token (default: line)
// into the provided channel.
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
		bufSize = 64 * 1024 // 64 KiB
	}

	maxToken := cfg.MaxToken
	if maxToken <= 0 {
		maxToken = bufSize
	}

	scanner.Buffer(make([]byte, bufSize), maxToken)
}

// CopyFile copies a file from src to dst.
// The destination file is replaced if it exists.
func CopyFile(src, dst string) error {
	if err := ensureDir(dst); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("create destination file %s: %w", dst, err)
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(dst)
		return fmt.Errorf("copy file %s -> %s: %w", src, dst, err)
	}

	if err := out.Sync(); err != nil {
		_ = out.Close()
		_ = os.Remove(dst)
		return err
	}

	return out.Close()
}

