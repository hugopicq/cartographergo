package cartographer

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Socket struct {
	IP   string
	Port uint16
}

type HostNameResult struct {
	Name string
	IP   []net.IP
}

func buildSockets(addresses []string, ports []uint16) []Socket {
	sockets := make([]Socket, len(addresses)*len(ports))
	//We build the list to iterate over port first, then on host to avoid scanning multiple ports on a single host at a time if possible
	for p, port := range ports {
		for a, address := range addresses {
			socket := Socket{
				IP:   address,
				Port: port,
			}
			sockets[p*len(addresses)+a] = socket
		}
	}
	return sockets
}

func ResolveComputersIP(computersByName map[string]*Computer, batchSize uint16, domain string) {
	totalProcessed := 0
	wg := sync.WaitGroup{}

	const ipPattern string = "^(?:[0-9]{1,3}\\.){3}[0-9]{1,3}$"

	hostnames := make([]string, 0, len(computersByName))
	for k, _ := range computersByName {
		hostnames = append(hostnames, k)
	}

	for {
		ch := make(chan HostNameResult)
		for i := 0; i < int(batchSize); i++ {
			if totalProcessed >= len(computersByName) {
				break
			}
			wg.Add(1)
			go func(ch chan<- HostNameResult, hostname string) {
				result := resolveComputerName(hostname, domain)
				ch <- result
				wg.Done()
			}(ch, hostnames[totalProcessed])
			totalProcessed += 1
		}

		go func() {
			wg.Wait()
			close(ch)
		}()

		for result := range ch {
			if len(result.IP) > 0 {
				ipaddr := ""
				for _, r := range result.IP {
					result, err := regexp.MatchString(ipPattern, r.String())
					if err == nil && result {
						//TODO : Handle multiple IP
						ipaddr = r.String()
						break
					}
				}
				currentcomputer := computersByName[result.Name]
				currentcomputer.IP = ipaddr
			}
		}

		if totalProcessed >= len(computersByName) {
			break
		}
	}
}

func GetComputersLDAP(cred *Credentials, includeWorkstations bool, ldaps bool) []Computer {
	filter := "(&(sAMAccountType=805306369)(operatingsystem=*Server*))"
	if includeWorkstations {
		filter = "(sAMAccountType=805306369)"
	}
	attributes := []string{"name", "dn", "operatingSystem", "userAccountControl", "ms-MCS-AdmPwd"}
	entries, err := ExecuteLDAPQuery(cred, filter, attributes, 256, ldaps)

	if err != nil {
		log.Fatal("Error while retrieving computers LDAP: " + err.Error())
	}

	computers := make([]Computer, len(entries))
	for c, computer := range entries {
		isDC := false
		uac, err := strconv.Atoi(computer.GetAttributeValue("userAccountControl"))
		if err == nil && (uac&8192 == 8192) {
			isDC = true
		}

		laps := computer.GetAttributeValue("ms-MCS-AdmPwd") != ""

		computers[c] = Computer{
			IP:              "",
			Name:            computer.GetAttributeValue("name"),
			OperatingSystem: computer.GetAttributeValue("operatingSystem"),
			IsDC:            isDC,
			LAPS:            laps,
			OpenPorts:       []uint16{},
			ModuleResults:   map[string]string{},
		}
	}

	return computers

}

func ScanPorts(hosts []string, ports []uint16, batchSize uint16, timeout time.Duration) []Socket {
	opensockets := []Socket{}
	sockets := buildSockets(hosts, ports)
	totalProcessed := 0
	wg := sync.WaitGroup{}

	for {
		ch := make(chan Socket)

		for i := 0; i < int(batchSize); i++ {
			if totalProcessed >= len(sockets) {
				break
			}
			wg.Add(1)
			go func(ch chan<- Socket, socket Socket) {
				opened := scanPort(socket.IP, socket.Port, timeout)
				if opened {
					ch <- socket
				}
				wg.Done()
			}(ch, sockets[totalProcessed])
			totalProcessed += 1
		}

		//One last go function that waits for all scans to finish before closing the channel
		go func() {
			wg.Wait()
			close(ch)
		}()

		for result := range ch {
			opensockets = append(opensockets, result)
		}

		if totalProcessed >= len(sockets) {
			break
		}
	}

	return opensockets
}

func scanPort(address string, port uint16, timeout time.Duration) bool {
	dialer := net.Dialer{Timeout: timeout}
	endpoint := address + ":" + strconv.Itoa(int(port))
	_, err := dialer.Dial("tcp", endpoint)
	if err != nil {
		return false
	}
	return true
}

func resolveComputerName(hostname string, domain string) HostNameResult {
	//TODO : Handle multiple addresses case
	hostnameDomain := hostname
	if strings.HasSuffix(strings.ToLower(hostnameDomain), strings.ToLower(domain)) == false {
		hostnameDomain = fmt.Sprintf("%v.%v", hostname, domain)
	}

	addr, err := net.LookupIP(hostnameDomain)
	if err == nil && len(addr) > 0 {
		return HostNameResult{Name: hostname, IP: addr}
	}
	return HostNameResult{Name: hostname, IP: []net.IP{}}
}
