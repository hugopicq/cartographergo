package cmd

import (
	"bufio"
	"encoding/csv"
	"log"
	"net"
	"os"
	"regexp"
	"time"

	"github.com/hugopicq/cartographergo/scanner"
	"github.com/spf13/cobra"
)

var dc string
var user string
var password string
var domain string
var whitelistfile string
var outputfile string
var batchsize uint16
var timeout uint

var rootCmd = &cobra.Command{
	Use:   "cartographer",
	Short: "A brief description of your application",
	Long:  `Pouvoir is a CLI application developped to perform reconnaissance on an internal network.`,
	Run:   main,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&dc, "domaincontroller", "s", "", "IP of Domain Controller")
	rootCmd.MarkFlagRequired("domaincontroller")
	rootCmd.Flags().StringVarP(&user, "user", "u", "", "Active Directory User")
	rootCmd.MarkFlagRequired("user")
	rootCmd.Flags().StringVarP(&password, "password", "p", "", "Active Directory Password")
	rootCmd.MarkFlagRequired("password")
	rootCmd.Flags().StringVarP(&domain, "domain", "d", "", "Active Directory Domain")
	rootCmd.MarkFlagRequired("domain")
	rootCmd.Flags().StringVarP(&outputfile, "outputfile", "o", "", "Output filepath")
	rootCmd.MarkFlagRequired("outputfile")

	rootCmd.Flags().StringVarP(&whitelistfile, "whitelistfile", "w", "", "Whitelist IP files in IP or CIDR format")
	rootCmd.Flags().Uint16VarP(&batchsize, "batchsize", "b", 4500, "Batch size")
	rootCmd.Flags().UintVarP(&timeout, "timeout", "t", 1500, "Timeout in milliseconds")

	//TODO : Add SSL support
}

func main(cmd *cobra.Command, args []string) {
	log.Println("Starting cartographer")

	whitelistIP := []string{}
	if whitelistfile != "" {
		log.Println("Processing whitelist file...")
		whitelistIP = readWhitelist()
	}

	log.Println("Getting computer information from DC...")
	computers := scanner.GetComputersLDAP(dc, user, password, domain)

	//From there we have the computer LIST

	//Convert to Hashmap
	computersByName := map[string]*scanner.Computer{}
	for k, computer := range computers {
		computersByName[computer.Name] = &computers[k]
	}

	log.Println("Resolving hostnames...")
	scanner.ResolveComputersIP(computersByName, batchsize)

	log.Println("Preparing scanning...")
	//Building hashmap and list of IP to scan
	ipToScan := []string{}
	computersByIP := map[string]*scanner.Computer{}
	for k, computer := range computers {
		computersByIP[computer.IP] = &computers[k]
		if computer.IP != "" && (len(whitelistIP) == 0 || contains(whitelistIP, computer.IP)) {
			ipToScan = append(ipToScan, computer.IP)
		}
	}

	log.Println("Starting port scan...")
	results := scanner.ScanPorts(ipToScan, scanner.TOP_1000_PORTS, batchsize, time.Millisecond*time.Duration(timeout))

	log.Println("Processing results...")
	for _, result := range results {
		computer := computersByIP[result.IP]
		(*computer).OpenPorts = append((*computer).OpenPorts, result.Port)
	}

	log.Println("Output result to file...")
	writeResults(computers, outputfile)

	log.Println("Done!")
}

func readWhitelist() []string {
	whitelist := []string{}

	file, err := os.Open(whitelistfile)
	if err != nil {
		log.Fatal("Problem while reading whitelist file: ", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		whitelist = append(whitelist, scanner.Text())
	}

	return processWhitelistCIDR(whitelist)
}

func writeResults(computers []scanner.Computer, outputfile string) {

	headers := []string{"Name", "IP", "OS", "OpenPorts"}

	f, err := os.Create(outputfile)
	if err != nil {
		log.Fatal("Failed to open output file: " + err.Error())
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.Comma = ';'
	defer w.Flush()

	if err := w.Write(headers); err != nil {
		log.Fatal("Error while writing headers to file: ", err)
	}

	for _, computer := range computers {
		if err := w.Write(computer.ToCSVLine()); err != nil {
			log.Fatal("Error writing record to file: ", err)
		}
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func processWhitelistCIDR(whitelistCIDR []string) []string {
	whitelistIP := []string{}
	const ipPattern string = "^(?:[0-9]{1,3}\\.){3}[0-9]{1,3}$"

	for _, ipcidr := range whitelistCIDR {
		matchIP, err := regexp.MatchString(ipPattern, ipcidr)
		if err != nil {
			//Should return error
			log.Fatal("Error while parsing IPCIDR: " + err.Error())
			return whitelistIP
		}

		if matchIP {
			//Then it is an IP
			whitelistIP = append(whitelistIP, ipcidr)
		} else {
			//Then we suppose it is a CIDR but TODO we should check for it
			ip, ipnet, err := net.ParseCIDR(ipcidr)
			if err != nil {
				//It is not IPCIDR
				log.Fatal("Error while parsing IPCIDR: " + err.Error())
				return whitelistCIDR
			}

			for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
				whitelistIP = append(whitelistIP, ip.String())
			}
		}
	}

	return whitelistIP
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
