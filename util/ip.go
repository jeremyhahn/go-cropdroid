package util

import (
	"net"
)

// parseLocalIP returns the first routable IP on the host
// or localhost if a LAN/WAN interface is not found.
func ParseLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			panic(err)
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	//return "localhost"
	panic("Couldn't find any routeable IP addresses")
}

func Nslookup(host string) (string, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return "", err
	}
	return ips[0].String(), nil
}
