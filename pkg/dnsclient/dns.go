package dnsclient

import (
	"fmt"
	"net"
	"strings"
)

type Client struct {
	Server string
}

func NewDNSClient(server string) *Client {
	return &Client{Server: server}
}

//GetAuthoritativeNS returns the first authoritative name server (from NS records) of a domain
func (c *Client) GetAuthoritativeNS(domain string) (string, error) {
	dparts := strings.Split(domain, ".")
	for i := 0; i < len(dparts) - 1 ; i++ {
		nss, err := net.LookupNS(strings.Join(dparts[i:], "."))
		if err != nil {
			continue
		}
		if len(nss) > 0 {
			return nss[0].Host+":53", nil
		}
	}
	return "", fmt.Errorf("No nameservers found for domain %s", domain)
}