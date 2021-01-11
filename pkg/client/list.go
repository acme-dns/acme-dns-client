package client

import (
	"fmt"
	"github.com/acme-dns/acme-dns-client/pkg/dnsclient"
)

func (c *AcmednsClient) List() {
	adnsAccts := c.Storage.FetchAll()
	functional := make([]string, 0)
	dysfunctional := make([]string, 0)
	errored := make([]string, 0)

	if len(adnsAccts) == 0 {
		fmt.Printf("No acme-dns accounts were found on this system.\n")
	} else {
		fmt.Printf("Number of acme-dns accounts found on this system: %d\nPerforming CNAME checks...\n\n", len(adnsAccts))
		for d, acct := range adnsAccts {
			dnsc := dnsclient.NewDNSClient(c.Config.DNSServer)
			cname, err := dnsc.GetCNAME(d)
			if err != nil {
				errored = append(errored, fmt.Sprintf("%s (%s)", d, err))
			} else if cname.CorrectTarget(acct.FullDomain) {
				functional = append(functional, d)
			} else {
				dysfunctional = append(dysfunctional, d)
			}
		}
	}
	if len(functional) > 0 {
		fmt.Printf("Working:\n")
		for _, s := range functional {
			PrintSuccess(s, 0)
		}
		fmt.Println()
	}
	if len(errored) > 0 {
		fmt.Printf("Error:\n")
		for _, e := range errored {
			PrintError(e, 0)
		}
		fmt.Println()
	}
	if len(dysfunctional) > 0 {
		fmt.Printf("Dysfunctional:\n")
		for _, d := range dysfunctional {
			PrintWarning(d, 0)
		}
	}
}
