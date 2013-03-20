package netutils

import (
	"net"
)

func IsInvalidAddress(addr string) bool {
	if 0 == len(addr) {
		return false
	}
	return IsInvalid(net.ParseIP(addr))
}

func IsInvalid(ip net.IP) bool {
	if nil == ip {
		return false
	}
	return ip.IsUnspecified() ||
		ip.IsLoopback() ||
		ip.IsInterfaceLocalMulticast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsMulticast()
}
