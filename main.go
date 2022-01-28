package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/acme-dns/acme-dns-client/pkg/client"
)

const (
	storagepath = "/etc/acmedns/clientstorage.json"
	VERSION     = "0.3"
)

func main() {
	conf := client.NewAcmednsConfig()
	flag.Usage = UsageGeneric

	checkFlags := flag.NewFlagSet("check", flag.ExitOnError)
	checkFlags.BoolVar(&conf.Verbose, "v", false, "Verbose output")
	checkFlags.BoolVar(&conf.Debug, "vv", false, "Very verbose (DEBUG) output")
	checkFlags.StringVar(&conf.DNSServer, "ns", "1.1.1.1:53", "Fallback DNS server and port to use for lookups")
	checkFlags.StringVar(&conf.Domain, "d", "", "Target domain name")

	checkFlags.Usage = FSUsage(checkFlags)

	registerFlags := flag.NewFlagSet("register", flag.ExitOnError)
	registerFlags.BoolVar(&conf.Dangerous, "dangerous", false, "Acknowledgement that this is a dangerous action")
	registerFlags.BoolVar(&conf.Verbose, "v", false, "Verbose output")
	registerFlags.BoolVar(&conf.Debug, "vv", false, "Very verbose (DEBUG) output")
	registerFlags.StringVar(&conf.DNSServer, "ns", "1.1.1.1:53", "Fallback DNS server and port to use for lookups")
	registerFlags.StringVar(&conf.Domain, "d", "", "Target domain name")
	registerFlags.StringVar(&conf.Server, "s",
		"https://auth.acme-dns.io", "Acme-dns server instance to use")
	registerFlags.StringVar(&conf.AllowList, "allow", "",
		"Comma separated allowlist of CIDR masks that are allowed use this acme-dns account. (Default: allow from all)")

	registerFlags.Usage = FSUsage(registerFlags)

	listFlags := flag.NewFlagSet("list", flag.ExitOnError)
	listFlags.BoolVar(&conf.Verbose, "v", false, "Verbose output")
	listFlags.BoolVar(&conf.Debug, "vv", false, "Very verbose (DEBUG) output")
	listFlags.StringVar(&conf.DNSServer, "ns", "1.1.1.1:53", "Fallback DNS server and port to use for lookups")

	listFlags.Usage = FSUsage(listFlags)

	// Server flag for validation
	flag.StringVar(&conf.Server, "s",
		"https://auth.acme-dns.io", "Acme-dns server instance to use")

	err := preflight()
	if err != nil {
		fmt.Printf("Error while starting up: %s\n", err)
		os.Exit(1)
	}
	// Preflight should have ensured that we have the storagepath structure created
	adnsClient := client.NewAcmednsClient(storagepath)
	adnsClient.Config = conf

	if len(os.Args) < 2 {
		flag.Parse()
		if !adnsClient.Validation() {
			UsageGeneric()
			os.Exit(1)
		}
		// Successfully validated
		os.Exit(0)
	}

	switch os.Args[1] {
	case "check":
		checkFlags.Parse(os.Args[2:])
		// Remove *. as the wildcard CNAME path is the same as the main domains
		conf.Domain = strings.Replace(conf.Domain, "*.", "", -1)
		adnsClient.CheckAndPrint()
	case "register":
		registerFlags.Parse(os.Args[2:])
		// Remove *. as the wildcard CNAME path is the same as the main domains
		conf.Domain = strings.Replace(conf.Domain, "*.", "", -1)
		adnsClient.Register()
	case "list":
		listFlags.Parse(os.Args[2:])
		adnsClient.List()
	default:
		// This handles --help, -h etc and if found, exits.
		flag.Parse()
		// We reach this only if no --help etc. was found
		if !adnsClient.Validation() {
			UsageGeneric()
			os.Exit(1)
		}
	}
}

func preflight() error {
	var err error
	if _, err = os.Stat(filepath.Dir(storagepath)); os.IsNotExist(err) {
		err = os.Mkdir(filepath.Dir(storagepath), 0700)
	}
	return err
}
