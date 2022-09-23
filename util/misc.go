package util

import (
	"net"
	"regexp"
	"strings"
)

func StringsContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func UInt16Contains(a []uint16, e uint16) bool {
	for _, b := range a {
		if b == e {
			return true
		}
	}
	return false
}

func StringsToString(ss []string) string {
	return "|" + strings.Trim(strings.Join(ss, "|"), "[]") + "|"
}

func CIDRToStrings(whitelistCIDR []string) ([]string, error) {
	whitelistIP := []string{}
	const ipPattern string = "^(?:[0-9]{1,3}\\.){3}[0-9]{1,3}$"

	for _, ipcidr := range whitelistCIDR {
		matchIP, err := regexp.MatchString(ipPattern, ipcidr)
		if err != nil {
			return whitelistIP, err
		}

		if matchIP {
			//Then it is an IP
			whitelistIP = append(whitelistIP, ipcidr)
		} else {
			//Then we suppose it is a CIDR but TODO we should check for it
			ip, ipnet, err := net.ParseCIDR(ipcidr)
			if err != nil {
				return whitelistIP, err
			}

			for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
				whitelistIP = append(whitelistIP, ip.String())
			}
		}
	}

	return whitelistIP, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
