package scanner

import (
	"bufio"
	"net"
	"os"
	"strings"
)

// StreamIPs reads IPs from a file line by line and sends them to a channel.
// If limit > 0, stops after sending limit IPs.
func StreamIPs(path string, limit int) (chan string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	ipChan := make(chan string, 100) // buffered to smooth out worker consumption

	go func() {
		defer file.Close()
		defer close(ipChan)

		scanner := bufio.NewScanner(file)
		sent := 0

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			ip := strings.Split(line, " ")[0]
			if ip == "" || !IsValidIP(ip) {
				continue
			}

			ipChan <- ip
			sent++

			if limit > 0 && sent >= limit {
				break
			}
		}
	}()

	return ipChan, nil
}

// IsValidIP returns true if the string is a valid IPv4 or IPv6 address.
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// CountValidIPs counts the number of valid IPs in the given file.
// If max > 0, stops counting once max is reached.
func CountValidIPs(path string, max int) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		ip := strings.Split(line, " ")[0]
		if ip == "" || !IsValidIP(ip) {
			continue
		}
		count++
		if max > 0 && count >= max {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return count, err
	}

	return count, nil
}
