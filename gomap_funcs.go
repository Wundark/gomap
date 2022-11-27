package gomap

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strings"
)

func canSocketBind(laddr string) bool {
	// Check if user can listen on socket
	listenAddr, err := net.ResolveIPAddr("ip4", laddr)
	if err != nil {
		return false
	}

	conn, err := net.ListenIP("ip4:tcp", listenAddr)
	if err != nil {
		return false
	}

	conn.Close()
	return true
}

// createHostRange converts a input ip addr string to a slice of ips on the cidr
func createHostRange(netw string) []string {
	_, ipv4Net, err := net.ParseCIDR(netw)
	if err != nil {
		log.Fatal(err)
	}

	mask := binary.BigEndian.Uint32(ipv4Net.Mask)
	start := binary.BigEndian.Uint32(ipv4Net.IP)
	finish := (start & mask) | (mask ^ 0xffffffff)

	var hosts []string
	for i := start + 1; i <= finish-1; i++ {
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, i)
		hosts = append(hosts, ip.String())
	}

	return hosts
}

// getLocalRanges returns local ip range or defaults on error to most common
func getLocalRanges() []string {
	defaultReturn := []string{"192.168.1.0/24"}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return defaultReturn
	}
	var allAddresses []string
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				split := strings.Split(ipnet.IP.String(), ".")
				mask, _ := ipnet.Mask.Size()
				// If a CIDR mask is smaller than 24, don't use a range to 0
				if mask <= 24 {
					allAddresses = append(allAddresses, fmt.Sprintf("%s.%s.%s.0/%d", split[0], split[1], split[2], mask))
				} else {
					allAddresses = append(allAddresses, address.String())
				}
			}
		}
	}
	if len(allAddresses) == 0 {
		return defaultReturn
	}
	return allAddresses
}

// getLocalIP returns local ip range or defaults on error to most common
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), err
			}
		}
	}
	return "", fmt.Errorf("No IP Found")
}

// getLocalIP returns local ip range or defaults on error to most common
func getLocalIPsForRanges() (map[string]string, error) {
	localRanges := getLocalRanges()

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	localRangesMap := map[string]string{}

	for _, localRange := range localRanges {
		for _, address := range addrs {
			// check the address type and if it is not a loopback the display it
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					_, ipNetRange, _ := net.ParseCIDR(localRange)
					if ipNetRange.Contains(ipnet.IP) {
						localRangesMap[localRange] = ipnet.IP.String()
					}
				}
			}
		}
	}

	if len(localRangesMap) == 0 {
		return nil, fmt.Errorf("No IPs Found")
	}

	return localRangesMap, nil
}
