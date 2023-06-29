package client

import (
	"fmt"
	"github.com/acme-dns/acme-dns-client/pkg/dnsclient"
	"github.com/acme-dns/acme-dns-client/pkg/integration"
)

var (
	CAA_SETTINGS = `Please copy the CAA record information for the ACME account you are using and add it to your
domain's DNS zone. CAA record with "issue" tag is used for exact domain names and "issuewild" for wildcard certificates.
You can add either or both of them based on your needs.

acme-dns-client will now proceed to check for the CAA records every 15 seconds and will continue after they're added
`
	CAA_INFO_ACCOUNT_NOTFOUND = `Could not find ACME accounts created by supported ACME clients on the system. Please add a CAA
record manually. For this you are going to need to look into how your client stores the ACME account URI, modify the 
example CAA record below accordingly and add it to your DNS zone:
---------------------------

%[1]s.         IN    CAA    0 issue "letsencrypt.org; validationmethods=dns-01; accounturi=https://acme-v01.api.letsencrypt.org/acme/reg/ACCOUNTUID"
%[1]s.         IN    CAA    0 issuewild "letsencrypt.org; validationmethods=dns-01; accounturi=https://acme-v01.api.letsencrypt.org/acme/reg/ACCOUNTUID"

---------------------------
`
)

func (c *AcmednsClient) CNAMESetupWizard(domain string) bool {
	c.Debug("Trying to fetch existing account for the domain from storage")
	acct, err := c.Storage.Fetch(c.Config.Domain)
	if err != nil {
		PrintError(fmt.Sprintf("Error while trying to fetch acme-dns account from storage: %s", err),0)
		return false
	}
	fmt.Printf(CNAME_INFO, domain, acct.FullDomain, domain, acct.FullDomain)
	c.Debug("Starting DNS monitoring for CNAME changes")
	return c.monitorCNAMERecordChange(domain, acct.FullDomain)
}

func (c *AcmednsClient) CAASetupWizard(domain string) bool {
	accts := c.findACMEAccounts()
	if len(accts) > 0 {
		PrintInfo(fmt.Sprintf("Found a total of %d ACME account(s) on this system:", len(accts)), 0)
		for _, a := range accts {
			recString, err := a.CAARecordString()
			if err != nil {
				c.Verbose(fmt.Sprintf("Error while generating CAA record string: %s", err))
			}
			fmt.Printf("  [%s] URI: %s\n", a.Client, a.URI)
			c.Verbose(fmt.Sprintf("  Contact: %s\n", a.Contact))
			c.Verbose(fmt.Sprintf("  Filepath: %s\n", a.FilePath))
			if recString != "" {
				fmt.Printf("  CAA record info:\n    -----------------------------------------------\n\n")
				fmt.Printf("    %s.             IN    CAA    0 issue %s\n", domain, recString)
				fmt.Printf("    %s.             IN    CAA    0 issuewild %s\n\n", domain, recString)
			}
			fmt.Printf("    -----------------------------------------------\n")
		}
		fmt.Printf(CAA_SETTINGS)
		return c.monitorCAARecordChange(domain)
	} else {
		fmt.Printf(CAA_INFO_ACCOUNT_NOTFOUND, domain)
		if YesNoPrompt("Do you want acme-dns-client to monitor for CAA record change?", false) {
			return c.monitorCAARecordChange(domain)
		}
		fmt.Printf(`After creation, the configuration for the domain %s by issuing the following command: 
    acme-dns-client check -d %s
`, domain, domain)
	}
	return false
}

func (c *AcmednsClient) monitorCAARecordChange(domain string) bool {
	fmt.Printf("Waiting for CAA record to be created for domain %s\n", domain)
	fmt.Printf("Querying the authoritative nameserver every 15 seconds.\n\n")
	dnsc := dnsclient.NewDNSClient(c.Config.DNSServer)
	for {
		newcaa, err := dnsc.GetCAA(domain)
		if err != nil && err != dnsclient.ErrCAARecordNotFound {
			PrintError(fmt.Sprintf("Caught an error while trying to query for CAA record: %s", err), 0)
			return false
		}
		for _, caa := range newcaa {
			if caa.IsSet() {
				c.Verbose(fmt.Sprintf("CAA record data: %s", caa.Data))
				PrintSuccess("Record found!", 0)
				return true
			}
		}
		printPauseCounter(15)
	}
	return false
}

func (c *AcmednsClient) monitorCNAMERecordChange(domain string, target string) bool {
	fmt.Printf("Waiting for CNAME record to be set up for domain %s\n", domain)
	fmt.Printf("Querying the authoritative nameserver every 15 seconds.\n\n")
	dnsc := dnsclient.NewDNSClient(c.Config.DNSServer)
	oldcname, err := dnsc.GetCNAME(domain)
	if err != nil && err != dnsclient.ErrCNAMERecordNotFound {
		PrintError(fmt.Sprintf("Caught an error while trying to query for CNAME record: %s", err), 0)
		return false
	}
	for {
		cname, err := dnsc.GetCNAME(domain)
		if err != nil && err != dnsclient.ErrCNAMERecordNotFound {
			PrintError(fmt.Sprintf("Caught an error while trying to query for CNAME record: %s", err), 0)
			return false
		}
		if cname.Target != oldcname.Target {
			c.Verbose(fmt.Sprintf("Detected a change in CNAME record. New CNAME target: %s", cname.Target))
			oldcname = cname
		}
		if cname.CorrectTarget(target) {
			PrintSuccess("CNAME record is now correctly set up!", 0)
			return true
		}
		printPauseCounter(15)
	}
	return false
}

func (c *AcmednsClient) findACMEAccounts() []integration.ACMEAccount {
	acmeAccts := make([]integration.ACMEAccount, 0)
	// Get accounts from integrations
	acmeclients := integration.GetIntegrations()
	for _, i := range acmeclients {
		if i.Found() {
			c.Verbose(fmt.Sprintf("Installation of ACME client %s found, looking for accounts", i.Name()))
			accts, err := i.FindAccounts()
			if err != nil {
				c.Debug(fmt.Sprintf("Error while looking for %s ACME accounts: %s", i.Name(), err))
			} else {
				c.Verbose(fmt.Sprintf("Found %d account(s)", len(accts)))
				acmeAccts = append(acmeAccts, accts...)
			}
		}
		c.Debug(fmt.Sprintf("Looking for ACME accounts from %s configuration", i.Name()))
	}
	return acmeAccts
}