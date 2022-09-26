package utils

import (
	"encoding/binary"
	"encoding/hex"
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

func BytesToHexString(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	stringdata := hex.EncodeToString(data)
	finaldata := ""
	for i := 0; i < len(stringdata); i += 2 {
		finaldata = finaldata + string(stringdata[i]) + string(stringdata[i+1]) + " "
	}
	return strings.ToUpper(finaldata[:len(finaldata)-1])
}

func BinToString(data []byte) string {
	uuid1 := hex.EncodeToString(binary.BigEndian.AppendUint32([]byte{}, binary.LittleEndian.Uint32(data[0:4])))
	uuid2 := hex.EncodeToString(binary.BigEndian.AppendUint16([]byte{}, binary.LittleEndian.Uint16(data[4:6])))
	uuid3 := hex.EncodeToString(binary.BigEndian.AppendUint16([]byte{}, binary.LittleEndian.Uint16(data[6:8])))
	uuid4 := hex.EncodeToString(data[8:10])
	uuid5 := hex.EncodeToString(data[10:12])
	uuid6 := hex.EncodeToString(data[12:16])

	return strings.ToUpper(uuid1 + "-" + uuid2 + "-" + uuid3 + "-" + uuid4 + "-" + uuid5 + uuid6)
}
