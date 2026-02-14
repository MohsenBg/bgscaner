package filemanager

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CSVConfig controls CSV read/write behavior
type CSVConfig struct {
	Comma            rune // field separator (default is ',')
	HasHeader        bool // skip the first row if true
	FieldsPerRecord  int  // -1 means variable fields
	LazyQuotes       bool
	TrimLeadingSpace bool
}

// StreamCSV reads CSV row-by-row and calls handler for each record
func StreamCSV(path string, cfg CSVConfig, handler func([]string) error) error {
	if !CheckFileExists(path) {
		return fmt.Errorf("file does not exist: %s", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open csv file %s: %w", path, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	applyCSVConfig(r, cfg)

	if cfg.HasHeader {
		_, _ = r.Read()
	}

	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("csv read error: %w", err)
		}
		if err := handler(rec); err != nil {
			return err
		}
	}

	return nil
}

// StreamCSVToChan sends each row into a channel
func StreamCSVToChan(path string, cfg CSVConfig, out chan<- []string) error {
	return StreamCSV(path, cfg, func(rec []string) error {
		out <- rec
		return nil
	})
}

// WriteCSVFile overwrites the file with all records
func WriteCSVFile(path string, cfg CSVConfig, records [][]string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create csv file %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	applyCSVConfigWriter(w, cfg)

	if err := w.WriteAll(records); err != nil {
		return fmt.Errorf("failed to write csv records: %w", err)
	}
	w.Flush()

	if err := w.Error(); err != nil {
		return fmt.Errorf("csv writer error: %w", err)
	}

	return nil
}

// StreamWriteCSV overwrites the file with all records
func StreamWriteCSV(
	path string,
	cfg CSVConfig,
	fn func(write func([]string) error) error,
) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create csv file %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	applyCSVConfigWriter(w, cfg)
	defer w.Flush()

	if err := fn(func(rec []string) error {
		return w.Write(rec)
	}); err != nil {
		return err
	}

	if err := w.Error(); err != nil {
		return fmt.Errorf("csv writer error: %w", err)
	}

	return nil
}

// AppendCSVRow appends a single row
func AppendCSVRow(path string, cfg CSVConfig, row []string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open csv file %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	applyCSVConfigWriter(w, cfg)

	if err := w.Write(row); err != nil {
		return fmt.Errorf("failed to append csv row: %w", err)
	}
	w.Flush()

	if err := w.Error(); err != nil {
		return fmt.Errorf("csv writer error: %w", err)
	}

	return nil
}

// AppendCSVRows appends multiple rows
func AppendCSVRows(path string, cfg CSVConfig, rows [][]string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open csv file %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	applyCSVConfigWriter(w, cfg)

	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return fmt.Errorf("failed to append csv row: %w", err)
		}
	}
	w.Flush()

	if err := w.Error(); err != nil {
		return fmt.Errorf("csv writer error: %w", err)
	}

	return nil
}

func applyCSVConfig(r *csv.Reader, cfg CSVConfig) {
	if cfg.Comma != 0 {
		r.Comma = cfg.Comma
	}
	if cfg.FieldsPerRecord != 0 {
		r.FieldsPerRecord = cfg.FieldsPerRecord
	} else {
		r.FieldsPerRecord = -1
	}
	r.LazyQuotes = cfg.LazyQuotes
}

func applyCSVConfigWriter(w *csv.Writer, cfg CSVConfig) {
	if cfg.Comma != 0 {
		w.Comma = cfg.Comma
	}
}
