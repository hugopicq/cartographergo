package cmd

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/hugopicq/cartographergo/cartographer"
	"github.com/hugopicq/cartographergo/cartographer/modules"
	"github.com/hugopicq/cartographergo/utils"
	"github.com/spf13/cobra"
)

var dc string
var user string
var password string
var domain string
var whitelistfile string
var blacklistfile string
var outputfile string
var batchsize uint16
var timeout uint
var includeWorkstations bool
var additionalIPString string
var skipAD bool
var cartoModules string
var runModulesAdditional bool
var ldaps bool

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
	// rootCmd.MarkFlagRequired("domaincontroller")
	rootCmd.Flags().StringVarP(&user, "user", "u", "", "Active Directory User")
	// rootCmd.MarkFlagRequired("user")
	rootCmd.Flags().StringVarP(&password, "password", "p", "", "Active Directory Password")
	// rootCmd.MarkFlagRequired("password")
	rootCmd.Flags().StringVarP(&domain, "domain", "d", "", "Active Directory Domain")
	// rootCmd.MarkFlagRequired("domain")
	rootCmd.Flags().StringVarP(&outputfile, "outputfile", "o", "", "Output filepath")
	rootCmd.MarkFlagRequired("outputfile")

	rootCmd.Flags().BoolVarP(&includeWorkstations, "include-workstations", "", false, "Include workstations for scans, by default only computers matching *Server* in the OS will be scanned")
	rootCmd.Flags().BoolVarP(&skipAD, "skip-domain", "", false, "Skip AD reconnaissance with modules to go straight to additional IP port scan")
	rootCmd.Flags().StringVarP(&whitelistfile, "whitelist-file", "", "", "Whitelist IP files in IP or CIDR format (1 line per CIDR)")
	rootCmd.Flags().StringVarP(&blacklistfile, "blacklist-file", "", "", "Blacklist IP files in IP or CIDR format (1 line per CIDR)")
	rootCmd.Flags().StringVarP(&additionalIPString, "additional-IP", "a", "", "Additional IP to scan. They must be in CIDR format separated by commas")
	rootCmd.Flags().StringVarP(&cartoModules, "modules", "m", "all", "Modules to run on resolved domain computers separated by commas (all by default). Available: all, adminsessions, webdav, shares, rpc, ftpanon")
	rootCmd.Flags().BoolVarP(&runModulesAdditional, "run-modules-additional", "", false, "Run chosen modules on additional IPs")
	rootCmd.Flags().BoolVarP(&ldaps, "ldaps", "", false, "Use LDAPS to communicate with DC (false by default)")
	rootCmd.Flags().Uint16VarP(&batchsize, "batchsize", "b", 4500, "Batch size")
	rootCmd.Flags().UintVarP(&timeout, "timeout", "t", 1500, "Timeout in milliseconds")

	//TODO : Add SSL support
}

func main(cmd *cobra.Command, args []string) {
	whitelistIP := []string{}
	blacklistIP := []string{}
	additionalIP := []string{}
	var err error

	if skipAD == false && (user == "" || password == "" || domain == "" || dc == "") {
		log.Fatal("If skip domain is not defined, the user, password, domain and dc arguments should be defined")
	}

	if whitelistfile != "" {
		log.Println("Processing whitelist file...")
		whitelistIP, err = readIPList(whitelistfile)
		if err != nil {
			log.Fatal("Problem while reading whitelist file: ", err)
		}
	}

	if blacklistfile != "" {
		log.Println("Processing blacklist file...")
		blacklistIP, err = readIPList(blacklistfile)
		if err != nil {
			log.Fatal("Problem while reading blacklist file: ", err)
		}
	}

	if additionalIPString != "" {
		additionalIP, err = parseAdditionalIP(additionalIPString)
		if err != nil {
			log.Fatal("Problem while parsing additional IPs: ", err)
		}
	}

	if runModulesAdditional && (user == "" || password == "" || domain == "" || dc == "") {
		log.Fatal("If modules are run on additional IP addresses, user, password, domain and dc arguments should be provided")
	}

	cartographer := cartographer.NewCartographer(dc, domain, user, password, batchsize, timeout, whitelistIP, blacklistIP, includeWorkstations, ldaps)

	cartoModulesArray := strings.Split(cartoModules, ",")

	if utils.StringsContains(cartoModulesArray, "all") || utils.StringsContains(cartoModulesArray, "shares") {
		cartographer.AddModule(new(modules.ModuleListShares), true)
	} else {
		cartographer.AddModule(new(modules.ModuleListShares), false)
	}

	if utils.StringsContains(cartoModulesArray, "all") || utils.StringsContains(cartoModulesArray, "ftpanon") {
		cartographer.AddModule(new(modules.ModuleFTPAnon), true)
	} else {
		cartographer.AddModule(new(modules.ModuleFTPAnon), false)
	}

	if utils.StringsContains(cartoModulesArray, "all") || utils.StringsContains(cartoModulesArray, "adminsessions") {
		cartographer.AddModule(modules.NewSessionsModule(ldaps), true)
	} else {
		cartographer.AddModule(modules.NewSessionsModule(ldaps), false)
	}

	if utils.StringsContains(cartoModulesArray, "all") || utils.StringsContains(cartoModulesArray, "webdav") {
		cartographer.AddModule(new(modules.ModuleWebDAV), true)
	} else {
		cartographer.AddModule(new(modules.ModuleWebDAV), false)
	}

	if utils.StringsContains(cartoModulesArray, "all") || utils.StringsContains(cartoModulesArray, "rpc") {
		cartographer.AddModule(new(modules.ModuleRPC), true)
	} else {
		cartographer.AddModule(new(modules.ModuleRPC), false)
	}

	if !skipAD || runModulesAdditional {
		cartographer.PrepareModules()
	}

	if !skipAD {
		cartographer.Run()
		cartographer.RunModules()
	}

	if len(additionalIP) > 0 {
		cartographer.RunAdditionalScan(additionalIP)
	}

	if runModulesAdditional {
		cartographer.RunModules()
	}

	log.Println("Output result to file...")
	err = cartographer.SaveResults(outputfile)
	if err != nil {
		log.Fatalf("Error writing results: %v", err)
	}

	log.Println("Done!")
}

func parseAdditionalIP(data string) ([]string, error) {
	ipCIDRList := strings.Split(data, ",")

	ipList, err := utils.CIDRToStrings(ipCIDRList)
	return ipList, err
}

func readIPList(filepath string) ([]string, error) {
	ipList := []string{}

	file, err := os.Open(filepath)
	if err != nil {
		return ipList, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ipList = append(ipList, scanner.Text())
	}

	ipList, err = utils.CIDRToStrings(ipList)

	return ipList, nil
}
