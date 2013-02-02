package netutils

import (
	"net"
)

func IsInvalidAddress(addr string) bool {
	return IsInvalid(net.ParseIP(addr))
}

func IsInvalid(ip net.IP) bool {
	return ip.IsUnspecified() ||
		ip.IsLoopback() ||
		ip.IsInterfaceLocalMulticast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsMulticast()
}
