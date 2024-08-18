package traefik_auth_with_exceptions

import (
	"fmt"
	"net"
	"time"
)

type ExceptionChecker struct {
	ipList             []net.IP
	ipNetList          []*net.IPNet
	hostList           []net.IP
	hostLastUpdate     time.Time
	hostUpdateInterval time.Duration
	rawHostList        []string
}

func NewExceptionChecker(config Exceptions) *ExceptionChecker {
	ipList, ipNetList := parseIpList(config.IpList)

	hostUpdateInterval, err := time.ParseDuration(config.HostUpdateInterval)
	if err != nil {
		fmt.Printf("Error parsing hostUpdateInterval: %v (host updates are disabled)\n", err)
		hostUpdateInterval = time.Duration(0)
	}

	ec := &ExceptionChecker{
		ipList:             ipList,
		ipNetList:          ipNetList,
		hostList:           []net.IP{},
		hostLastUpdate:     time.Time{},
		hostUpdateInterval: hostUpdateInterval,
		rawHostList:        config.HostList,
	}

	ec.UpdateHosts()

	return ec
}

func (ec *ExceptionChecker) UpdateHosts() {
	if len(ec.rawHostList) == 0 {
		return
	}

	timeSinceLastUpdate := time.Now().Sub(ec.hostLastUpdate)
	if ec.hostUpdateInterval <= 0 || timeSinceLastUpdate < ec.hostUpdateInterval {
		return
	}

	ec.hostList = resolveHosts(ec.rawHostList)
	ec.hostLastUpdate = time.Now()
}

func (ec *ExceptionChecker) IsTrustedRemoteAddr(addr string) bool {
	ec.UpdateHosts()

	ipRemote := splitAddr(addr)
	if ipRemote == nil {
		return false
	}

	for _, ip := range ec.ipList {
		if ip.Equal(ipRemote) {
			return true
		}
	}

	for _, ip := range ec.hostList {
		if ip.Equal(ipRemote) {
			return true
		}
	}

	for _, ipNet := range ec.ipNetList {
		if ipNet.Contains(ipRemote) {
			return true
		}
	}

	return false
}

func parseIpList(list []string) ([]net.IP, []*net.IPNet) {
	var ipList []net.IP
	var ipNetList []*net.IPNet

	for _, line := range list {
		if ip := parseIP(line); ip != nil {
			ipList = append(ipList, ip)
		} else if ipNet := parseIPNet(line); ipNet != nil {
			ipNetList = append(ipNetList, ipNet)
		} else {
			fmt.Printf("Parsing IP or CIDR failed: %s\n", line)
		}
	}

	return ipList, ipNetList
}

func parseIP(s string) net.IP {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil
	}

	return ip.To4()
}

func parseIPNet(s string) *net.IPNet {
	_, ipNet, err := net.ParseCIDR(s)
	if err != nil || ipNet.IP.To4() == nil {
		return nil
	}

	return ipNet
}

func lookupIP(s string) net.IP {
	ipList, err := net.LookupIP(s)
	if err != nil {
		fmt.Printf("Lookup IP failed for '%s': %v\n", s, err)
		return nil
	}

	for _, ip := range ipList {
		if ip.To4() != nil {
			return ip
		}
	}

	fmt.Printf("Lookup IP failed for '%s': no IPv4 found\n", s)
	return nil
}

func resolveHosts(list []string) []net.IP {
	var hostList []net.IP
	for _, item := range list {
		if ip := lookupIP(item); ip != nil {
			hostList = append(hostList, ip)
		}
	}

	return hostList
}

func splitAddr(addr string) net.IP {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		fmt.Printf("Splitting address '%s' failed: %v\n", addr, err)
		return nil
	}

	ip := parseIP(host)
	if ip == nil {
		fmt.Printf("Splitting address '%s' failed: no IP found\n", addr)
	}

	return ip
}
