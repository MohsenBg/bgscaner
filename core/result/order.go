package result

func (a ResultIPScan) Less(b ResultIPScan) bool {
	if a.Latency != b.Latency {
		return a.Latency < b.Latency
	}
	return a.IP < b.IP
}
