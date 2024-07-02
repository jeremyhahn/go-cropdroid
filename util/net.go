package util

import (
	"log"
	"net"
)

// Performs a logical UDP connection to an endpoint
// that does not need to exist and returns the local
// IP address that was used to perform the connection.
func PreferredLocalAddress() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}
