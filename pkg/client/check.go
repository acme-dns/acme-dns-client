package client

import (
	"fmt"
	"strings"

	"github.com/acme-dns/acme-dns-client/pkg/dnsclient"

	"github.com/cpu/goacmedns"
)

type ConfigurationState struct {
	Domain string
	Account goacmedns.Account
	CNAME dnsclient.CNAMERecord
	CAA []dnsclient.CAARecord
}

func NewConfigurationState(domain string) ConfigurationState {
	return ConfigurationState{
		Domain: domain,
		Account: goacmedns.Account{},
		CNAME: dnsclient.CNAMERecord{},
		CAA: make([]dnsclient.CAARecord, 0),
	}
}

func (c *AcmednsClient) CheckAndPrint() {
	domains := make([]string, 0)
	if c.Config.Domain != "" {
		// Prepare CLI provided domain list
		allDomains := strings.Split(c.Config.Domain, ",")
		for _, d := range allDomains {
			domains = append(domains, strings.TrimSpace(d))
		}
	} else {
		// Fetch all domains from storage
		for d, _ := range c.Storage.FetchAll() {
			domains = append(domains, d)
		}
	}
	for _, d := range domains {
		// Perform the check for each domain listed
		c.checkAndPrint(c.ConfigurationState(d))
	}
}

func (c *AcmednsClient) ConfigurationState(domain string) ConfigurationState {
	cstate := NewConfigurationState(domain)
	dnsc := dnsclient.NewDNSClient(c.Config.DNSServer)
	var err error

	// Populate CNAME record information
	cstate.CNAME, err = dnsc.GetCNAME(domain)
	if err != nil {
		c.Verbose(fmt.Sprintf("%s", err))
	}

	// Populate CAA record information
	cstate.CAA, err = dnsc.GetCAA(domain)
	if err != nil {
		c.Verbose(fmt.Sprintf("%s", err))
	}

	// Populate existing acme-dns account information
	cstate.Account, err = c.acmeDnsAccountForDomain(domain)
	if err != nil {
		c.Verbose(fmt.Sprintf("%s", err))
	}
	return cstate
}

// acmeDnsAccountForDomain returns a preconfigured `goacmedns.Account` for a domain or
// a fresh `goacmedns.Account` object if not found.
func (c *AcmednsClient) acmeDnsAccountForDomain(domain string) (goacmedns.Account, error) {
	adnsacct, err := c.Storage.Fetch(domain)
	if err != nil && err != goacmedns.ErrDomainNotFound{
		return goacmedns.Account{}, err
	} else if err == goacmedns.ErrDomainNotFound {
		return goacmedns.Account{}, nil
	}
	return adnsacct, err
}

func (c *AcmednsClient) checkAndPrint(cstate ConfigurationState) {
	fmt.Printf("Checking acme-dns configuration for domain %s\n", cstate.Domain)
	// Check acme-dns account and CNAME records
	c.PrintAcmednsAccountInfo(cstate)
	// Check CAA records
	cstate.PrintCAAResults()
}

func (c *AcmednsClient) PrintAcmednsAccountInfo(cstate ConfigurationState) {
	if cstate.Account.FullDomain == "" {
		PrintError("No acme-dns account registered", 1)
	} else {
		PrintSuccess("Registered acme-dns account found!",1)
		if cstate.CNAME.CorrectTarget(cstate.Account.FullDomain) {
			PrintSuccess("CNAME record found and set up correctly!", 1)
		} else if cstate.CNAME.Target != "" {
			PrintError(fmt.Sprintf("CNAME record found, but it's pointing to a wrong domain. expected: %s, found: %s",
				cstate.Account.FullDomain, cstate.CNAME.Target), 1)
			PrintInfo(fmt.Sprintf(`A correctly set up CNAME record should look like the following:
    _acme-challenge.%s. 120 IN      CNAME   %s.`, cstate.Domain, cstate.Account.FullDomain), 1)
		} else {
			PrintError(fmt.Sprintf("No CNAME record found"), 1)
			PrintInfo(fmt.Sprintf(`A correctly set up CNAME record should look like the following:
    _acme-challenge.%s.    IN      CNAME   %s.`, cstate.Domain, cstate.Account.FullDomain), 1)
			if YesNoPrompt("Do you want to set up the CNAME record now and have acme-dns-client monitor the change?", false) {
				_ = c.CNAMESetupWizard(cstate.Domain)
			}
		}
	}
}

func (c *ConfigurationState) PrintCAAResults() {
	if c.HasCAA() {
		fmt.Printf(" %s CAA record found!\n", successMarker())
	} else {
		fmt.Printf(" %s No CAA record found\n", warningMarker())
	}
	if c.HasAccountURI() {
		fmt.Printf(" %s CAA AccountURI found!\n", successMarker())
	} else {
		fmt.Printf(" %s No CAA AccountURI found\n", warningMarker())
	}
}

func (c *ConfigurationState) HasCAA() bool {
	for _, r := range c.CAA {
		if r.IsSet() {
			 return true
		}
	}
	return false
}

func (c *ConfigurationState) HasAccountURI() bool {
	for _, r := range c.CAA {
		if r.HasAccountURI() {
			return true
		}
	}
	return false
}

func (c *ConfigurationState) HasAcmednsAccount() bool {
	if c.Account.FullDomain != "" {
		return true
	}
	return false
}

func (c *ConfigurationState) CorrectCNAME() bool {
	if c.HasAcmednsAccount() && c.CNAME.CorrectTarget(c.Account.FullDomain) {
		return true
	}
	return false
}