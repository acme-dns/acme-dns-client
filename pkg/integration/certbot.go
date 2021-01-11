package integration

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

type CertbotAccount struct {
	TermsOfService string `json:"terms_of_service"`
	URI            string `json:"uri"`
	Body           struct {
		Contact []string `json:"contact"`
		Status  string   `json:"status"`
		Key     struct {
			E   string `json:"e"`
			Kty string `json:"kty"`
			N   string `json:"n"`
		} `json:"key"`
		Agreement string `json:"agreement"`
	} `json:"body"`
	NewAuthzrURI string `json:"new_authzr_uri"`
}

type CertbotClient struct {
	ConfigRoot string
}

func (c *CertbotClient) String() string {
	return "PLACEHOLDER"
}

func (c *CertbotClient) Name() string {
	return "Certbot"
}

//NewCertbotClient returns a new CertbotClient instance
func NewCertbotClient() *CertbotClient {
	return &CertbotClient{ConfigRoot: "/etc/letsencrypt"}
}

//Found checks if Certbot installation is found on the system, and config path
func (c *CertbotClient) Found() bool {
	if _, err := os.Stat(c.ConfigRoot); !os.IsNotExist(err) {
		return true
	}
	return false
}

//FindAccounts tries to search through Certbot configuration directory and to find ACME accounts registered using it
func (c *CertbotClient) FindAccounts() ([]ACMEAccount, error) {
	var accounts = make([]ACMEAccount, 0)
	err := filepath.Walk(path.Join(c.ConfigRoot, "accounts"),
		func(pth string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
 			if !info.IsDir() && info.Name() == "regr.json" {
 				newAcc, err := c.ParseAccountFile(pth)
 				if err != nil {
 					return err
				}
 				accounts = append(accounts, newAcc)
			}
			return err
		})
	return accounts, err
}

//ParseAccountFile parses Certbot account to a ACMEAccount struct
func (c *CertbotClient) ParseAccountFile(pth string) (ACMEAccount, error) {
	var err error
	cbacc := CertbotAccount{}
	acmeacc := ACMEAccount{
		FilePath: pth,
		Client: "Certbot",
	}

	data, err := ioutil.ReadFile(pth)
	if err != nil {
		return acmeacc, err
	}
	err = json.Unmarshal(data, &cbacc)
	if err != nil {
		return acmeacc, err
	}
	if len(cbacc.Body.Contact) > 0 {
		acmeacc.Contact = cbacc.Body.Contact[0]
	}
	acmeacc.URI = cbacc.URI
	return acmeacc, err
}

// FindValidationToken attempts to find a ACME validation token. For Certbot this is distributed
// via environmental variable CERTBOT_VALIDATION
func (c *CertbotClient) FindValidationToken() (string, error) {
	return os.Getenv("CERTBOT_VALIDATION"), nil
}

// FindValidationDomain attempts to find the domain the ACME validation is going to be carried for.
// For Certbot this is distributed via environmental variable CERTBOT_DOMAIN
func (c *CertbotClient) FindValidationDomain() (string, error) {
	return os.Getenv("CERTBOT_DOMAIN"), nil
}