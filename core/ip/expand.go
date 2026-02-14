package ip

import "net"

func StreamCIDR(cidr string, out chan<- string) error {
	ip, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}

	for curr := masked(ip, network.Mask); network.Contains(curr); curr = increment(curr) {
		out <- curr.String()
	}
	return nil
}

func increment(ip net.IP) net.IP {
	next := make(net.IP, len(ip))
	copy(next, ip)
	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] != 0 {
			break
		}
	}
	return next
}

func masked(ip net.IP, mask net.IPMask) net.IP {
	m := ip.Mask(mask)
	out := make(net.IP, len(m))
	copy(out, m)
	return out
}
