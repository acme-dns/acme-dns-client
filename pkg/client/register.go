package client

import (
	"fmt"
	"os"
	"strings"

	"github.com/cpu/goacmedns"
)

const PUBLIC_ACME_DNS = "https://auth.acme-dns.io"

var (
	CNAME_INFO = `
To finalize the setup, you need to create a CNAME record pointing from _acme-challenge.%s 
to the newly created acme-dns domain %s

A correctly set up CNAME record should look like the following:

_acme-challenge.%s.     IN      CNAME   %s.

`
	CHECK_INFO = `
After setting up the CNAME record to your main DNS zone, you can use acme-dns-client to check the configuration.
This can be done by issuing command:
  acme-dns-client check -d %s

`
	PUBLIC_INSTANCE_WARNING = `You are about to register an account to a public acme-dns instance. 
It's important to understand that using a third-party hosted acme-dns instance will authorize the acme-dns instance owner 
to request and validate certificates for your domain on your behalf. While the domain name is not known to acme-dns, it's 
trivial to deduct via correlation analysis with Certificate Transparency logs.

This issue can be mitigated in the future through CA's implementing support for ACME-CAA accounturi (RFC 8657), and
acme-dns-client will suggest you to set it up already. This is however optional.

In case this is intentional, and you understand the associated risks, re-run acme-dns-client with argument "--dangerous"
to suppress this this warning.
`
	CAA_INFO = `
A CAA record allows you to control additional certificate issuance safeguards. The currently supported
version allows the domain owner to control which certificate authorities are allowed to issue certificates for the domain in question.
The certificate authorities MUST check and respect the CAA records in the validation process.

There's also a standard (RFC 8657) that extends the CAA record to limit the issuance of certificates to a specific validation 
method and/or to a specific ACME account. While they can be tested using staging environment of Let's Encrypt for example,
they're not enabled in the production yet. It is still be worthwhile to configure them so you'll be protected when the feature gets enabled.
`
)

func (c *AcmednsClient) Register() {
	cstate := c.ConfigurationState(c.Config.Domain)
	if !c.Config.Dangerous && c.Config.Server == PUBLIC_ACME_DNS && !cstate.HasAcmednsAccount() {
		PrintWarning(fmt.Sprintf(PUBLIC_INSTANCE_WARNING), 0)
		os.Exit(0)
		/* TODO: enable when ACME-CAA hits production
		if cstate.HasCAA() {
			c.Verbose("CAA record found")
			if !cstate.HasAccountURI() {
				c.Verbose("CAA record does not have accounturi defined")
				// No CAA record with accountUri was found. We need to ensure that the user knows what they're doing
				if YesNoPrompt("Do you wish to set up a CAA record now?", true) {
					c.CAASetupWizard(c.Config.Domain)
				} else {
					os.Exit(0)
				}
			}
		} else {
			c.Verbose("No CAA record found")
			if YesNoPrompt("Do you wish to set up a CAA record now?", true) {
				c.CAASetupWizard(c.Config.Domain)
			} else {
				os.Exit(0)
			}
		}*/
	}

	c.Debug("Initializing goacmedns client")
	client := goacmedns.NewClient(c.Config.Server)

	c.Debug("Trying to fetch existing account for the domain from storage")
	if cstate.HasAcmednsAccount() {
		// TODO: ask if user wants to overwrite the existing and proceed with registration?
		PrintWarning(fmt.Sprintf("Acme-dns account already registered for domain %s", c.Config.Domain), 0)
	} else {
		// register a new account
		allowFrom := strings.Split(c.Config.AllowList, ",")
		c.Debug("Registering new account with the acme-dns server")
		newAccount, err := client.RegisterAccount(allowFrom)
		if err != nil {
			PrintError(fmt.Sprintf("%s", err), 0)
			return
		}

		cstate.Account = newAccount
		c.Debug("Adding the registered acme-dns account to storage state")
		err = c.Storage.Put(c.Config.Domain, cstate.Account)
		if err != nil {
			PrintError(fmt.Sprintf("%s", err), 0)
			return
		}

		c.Debug("Saving the acme-dns account storage to disk")
		err = c.Storage.Save()
		if err != nil {
			PrintError(fmt.Sprintf("%s", err), 0)
			return
		}
		PrintSuccess(fmt.Sprintf("New acme-dns account for domain %s successfully registered!\n", c.Config.Domain), 0)
	}

	if cstate.CorrectCNAME() {
		PrintSuccess("CNAME record seems to already be set up correctly, you are good to go", 0)
	} else {
		// Ask if user wants acme-dns-client to monitor CNAME change
		if YesNoPrompt("Do you want acme-dns-client to monitor the CNAME record change?", true) {
			c.CNAMESetupWizard(c.Config.Domain)
		} else {
			// if not, print post-check instruction
			c.PrintRegistrationInfo(c.Config.Domain, cstate.Account)
			fmt.Printf(CHECK_INFO, c.Config.Domain)
		}
	}

	if cstate.HasCAA() {
		c.Verbose("CAA record for the domain exists")
		if cstate.HasAccountURI() {
			c.Verbose("CAA accounturi for the domain exists")
		} else {
			fmt.Printf(CAA_INFO)
			if YesNoPrompt("Do you wish to set up a CAA record with accounturi now?", false) {
				c.CAASetupWizard(c.Config.Domain)
			}
		}
	} else {
		fmt.Printf(CAA_INFO)
		if YesNoPrompt("Do you wish to set up a CAA record now?", false) {
			c.CAASetupWizard(c.Config.Domain)
		}
	}
}

func (c *AcmednsClient) PrintRegistrationInfo(domain string, account goacmedns.Account) {
	fmt.Printf("Domain:         %s\n", account.FullDomain)
	c.Verbose(fmt.Sprintf("Username:   %s", account.Username))
	c.Verbose(fmt.Sprintf("Password:   %s", account.Password))
	fmt.Printf(CNAME_INFO, domain, account.FullDomain, domain, account.FullDomain)
}