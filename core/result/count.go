package result

func CountCSV(path string) (int64, error) {
	var n int64
	err := ReadCSV(path, func(_ ResultIPScan) error {
		n++
		return nil
	})
	return n, err
}
