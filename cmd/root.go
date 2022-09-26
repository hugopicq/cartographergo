package cmd

import (
	"bufio"
	"log"
	"os"

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
	whitelistIP := []string{}
	var err error
	if whitelistfile != "" {
		log.Println("Processing whitelist file...")
		whitelistIP, err = readWhitelist()
		if err != nil {
			log.Fatal("Problem while reading whitelist file: ", err)
		}
	}

	cartographer := cartographer.NewCartographer(dc, domain, user, password, batchsize, timeout, whitelistIP)
	cartographer.AddModule(new(modules.ModuleListShares))
	cartographer.AddModule(new(modules.SessionsModule))
	cartographer.AddModule(new(modules.ModuleWebDAV))
	cartographer.AddModule(new(modules.ModuleRPC))
	cartographer.Run()

	log.Println("Output result to file...")
	err = cartographer.SaveResults(outputfile)
	if err != nil {
		log.Fatal("Error writing results: %v", err)
	}

	log.Println("Done!")
}

func readWhitelist() ([]string, error) {
	whitelist := []string{}

	file, err := os.Open(whitelistfile)
	if err != nil {
		return whitelist, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		whitelist = append(whitelist, scanner.Text())
	}

	whitelist, err = utils.CIDRToStrings(whitelist)

	return whitelist, nil
}
