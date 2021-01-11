package integration

import (
	"fmt"
	"net/url"
	"strings"
)

type ACMEAccount struct {
	URI string
	Contact string
	Client string
	FilePath string
}

func (c *ACMEAccount) CAARecordString() (string, error) {
	// Get issuer from URI
	acctURL, err := url.Parse(c.URI)
	if err != nil {
		return "", err
	}
	hostparts := strings.Split(acctURL.Host, ".")
	if len(hostparts) > 1 {
		cadomain := strings.Join(hostparts[len(hostparts)-2:], ".")
		return fmt.Sprintf("\"%s; validationmethods=dns-01; accounturi=%s\"", cadomain, c.URI), nil
	}
	return "", fmt.Errorf("Encountered an error while trying to determine issuer domain from account URI host: %s", acctURL.Host)
}

func GetIntegrations() []ACMEClient {
	// Only Certbot supported right now
	integrations := make([]ACMEClient, 0)
	integrations = append(integrations, NewCertbotClient())
	return integrations
}