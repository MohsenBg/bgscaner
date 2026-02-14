package iplist

import (
	"strconv"
	"strings"

	"bgscan/core/ip"
)

func ParseCSV(rec []string) (IPList, bool) {
	if len(rec) == 0 {
		return IPList{}, false
	}

	raw := strings.TrimSpace(rec[0])
	ipStr, ok := ip.NormalizeIPOrCIDR(raw)
	if !ok {
		return IPList{}, false
	}

	enable := 1
	if len(rec) > 1 {
		if v, err := strconv.Atoi(strings.TrimSpace(rec[1])); err == nil {
			enable = v
		}
	}

	return New(ipStr, enable), true
}
