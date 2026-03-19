package result

// Less compares two scan results by latency, then by IP
func (a IPScanResult) Less(b IPScanResult) bool {
	if a.Download != b.Download {
		return a.Download < b.Download
	}
	if a.Upload != b.Upload {
		return a.Upload < b.Upload
	}
	if a.Latency != b.Latency {
		return a.Latency < b.Latency
	}
	return a.IP < b.IP
}

// Equal compares two scan results by IP
func (a IPScanResult) Equal(b IPScanResult) bool {
	return a.IP == b.IP
}
