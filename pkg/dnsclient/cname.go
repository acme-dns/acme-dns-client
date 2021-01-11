package dnsclient

import (
	"fmt"
	"github.com/miekg/dns"
)

var (
	ErrCNAMERecordNotFound = fmt.Errorf("No CNAME record found")
)

type CNAMERecord struct {
	Domain string
	HasCNAME bool
	Target string
}

func NewCNAMERecord() CNAMERecord {
	return CNAMERecord{
		Domain: "",
		HasCNAME: false,
		Target: "",
	}
}

// CorrectTarget returns true if the CNAME has been set correctly according to
// acme-dns account.
func (r *CNAMERecord) CorrectTarget(acmednsdomain string) bool {
	return dns.Fqdn(acmednsdomain) == dns.Fqdn(r.Target)
}

//GetCNAME fetches the CNAME for ACME "magic" subdomain _acme-challenge for a domain
func (c *Client) GetCNAME(domain string) (CNAMERecord, error) {
	domain = "_acme-challenge." + domain
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeCNAME)
	msg.RecursionDesired = true
	ns, err := c.GetAuthoritativeNS(domain)
	if err != nil {
		// Fallback to default nameserver
		ns = c.Server
	}
	in, err := dns.Exchange(msg, ns)
	if err != nil {
		return NewCNAMERecord(), err
	}
	for _, a := range in.Answer {
		if cname, ok := a.(*dns.CNAME); ok {
			return CNAMERecord{
				Domain: dns.Fqdn(domain),
				HasCNAME: true,
				Target: dns.Fqdn(cname.Target),
			}, nil
		} else {
			return NewCNAMERecord(), fmt.Errorf("Unexpected record returned with CNAME query to domain %s\n", domain)
		}
	}
	return NewCNAMERecord(), ErrCNAMERecordNotFound
}