package result

// Count returns number of valid records in CSV file
func Count(path string) (int64, error) {
	var count int64

	err := ReadCSV(path, func(_ IPScanResult) error {
		count++
		return nil
	})

	return count, err
}
