package ip

import (
	"fmt"
	"net"
)

func Count(input string) (uint64, error) {
	if _, netw, err := net.ParseCIDR(input); err == nil {
		ones, bits := netw.Mask.Size()
		return 1 << uint(bits-ones), nil
	}

	if net.ParseIP(input) != nil {
		return 1, nil
	}

	return 0, fmt.Errorf("invalid IP or CIDR: %s", input)
}
