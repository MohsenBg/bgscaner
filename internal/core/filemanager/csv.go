package filemanager

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ═══════════════════════════════════════════════════════════
// CSV Files Operations
// ═══════════════════════════════════════════════════════════

// CSVConfig controls CSV reader and writer behavior.
type CSVConfig struct {
	HasHeader        bool // skip the first row when reading
	LazyQuotes       bool // allow malformed quoting
	TrimLeadingSpace bool // trim leading spaces in fields
	Comma            rune // field separator (default ',')
	FieldsPerRecord  int  // -1 allows variable number of fields
}

// StreamCSV reads a CSV file row-by-row and calls handler for each record.
// It is memory efficient and suitable for very large CSV files.
func StreamCSV(path string, cfg CSVConfig, handler func([]string) error) error {
	if !CheckFileExists(path) {
		return fmt.Errorf("csv file does not exist: %s", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open csv file %s: %w", path, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	applyCSVConfig(r, cfg)

	if cfg.HasHeader {
		if _, err := r.Read(); err != nil && err != io.EOF {
			return fmt.Errorf("read csv header: %w", err)
		}
	}

	for {
		rec, err := r.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("csv read error: %w", err)
		}

		if err := handler(rec); err != nil {
			return err
		}
	}
}

// StreamCSVToChan reads a CSV file and sends each record to a channel.
func StreamCSVToChan(path string, cfg CSVConfig, out chan<- []string) error {
	return StreamCSV(path, cfg, func(rec []string) error {
		out <- rec
		return nil
	})
}

// WriteCSVFile overwrites the CSV file with all provided records.
func WriteCSVFile(path string, cfg CSVConfig, records [][]string) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create csv file %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	applyCSVConfigWriter(w, cfg)

	if err := w.WriteAll(records); err != nil {
		return fmt.Errorf("write csv records: %w", err)
	}

	w.Flush()
	return w.Error()
}

// StreamWriteCSV allows streaming CSV writing using a callback writer.
// Useful for writing large datasets without holding them in memory.
func StreamWriteCSV(
	path string,
	cfg CSVConfig,
	fn func(write func([]string) error) error,
) error {

	if err := ensureDir(path); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create csv file %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	applyCSVConfigWriter(w, cfg)

	write := func(rec []string) error {
		return w.Write(rec)
	}

	if err := fn(write); err != nil {
		return err
	}

	w.Flush()
	return w.Error()
}

// AppendCSVRow appends a single row to the CSV file.
func AppendCSVRow(path string, cfg CSVConfig, row []string) error {
	return AppendCSVRows(path, cfg, [][]string{row})
}

// AppendCSVRows appends multiple rows to the CSV file.
func AppendCSVRows(path string, cfg CSVConfig, rows [][]string) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open csv file %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	applyCSVConfigWriter(w, cfg)

	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return fmt.Errorf("append csv row: %w", err)
		}
	}

	w.Flush()
	return w.Error()
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
	r.TrimLeadingSpace = cfg.TrimLeadingSpace
}

func applyCSVConfigWriter(w *csv.Writer, cfg CSVConfig) {
	if cfg.Comma != 0 {
		w.Comma = cfg.Comma
	}
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}
