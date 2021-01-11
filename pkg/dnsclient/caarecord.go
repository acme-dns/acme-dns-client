package dnsclient

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

var (
	ErrCAARecordNotFound = fmt.Errorf("No CAA record found")
)

type CAARecord struct {
	Tag string
	Issuer string
	ValidationMethods []string
	AccountUri	string
	Data string
}

//NewRecord creates a new Record instance
func NewRecord() CAARecord {
	return CAARecord{
		Tag: "",
		Issuer: "",
		ValidationMethods: make([]string, 0),
		AccountUri: "",
	}
}

type CAACheckResult struct {
	HasCAA bool
	HasAccountUri bool
}

// HasAccountURI returns true if AccountURI for the CAA record has been set
func (c *CAARecord) HasAccountURI() bool {
	return len(c.AccountUri) > 0
}

// IsSet returns true if CAA record has been set. This is determined by if Issuer field exists
func (c *CAARecord) IsSet() bool {
	return len(c.Issuer) > 0
}

//CheckCAA performs checks to CAA records in order to determine if the domain has a CAA record and if the CAA
//record includes AccountURI parameter.
func (c *Client) CheckCAA(domain string) (CAACheckResult, error) {
	var check = CAACheckResult{}
	records, err := c.GetCAA(domain)
	if err != nil {
		return check, err
	}
	if len(records) > 0 {
		check.HasCAA = true
	}
	for _, r := range records  {
		if r.HasAccountURI() {
			check.HasAccountUri = true
		}
	}
	return check, nil
}

//GetCAA fetches the CAA records for a domain
func (c *Client) GetCAA(domain string) ([]CAARecord, error) {
	records := []CAARecord{}
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeCAA)
	msg.RecursionDesired = true

	ns, err := c.GetAuthoritativeNS(domain)
	if err != nil {
		// Fallback to default nameserver
		ns = c.Server
	}
	in, err := dns.Exchange(msg, ns)

	if err != nil {
		return records, err
	}

	for _, a := range in.Answer {
		if caa, ok := a.(*dns.CAA); ok {
			rec, err := ParseNewRecord(caa)
			if err != nil {
				return records, fmt.Errorf("Encountered an error while trying to parse CAA record: %s", err)
			}
			records = append(records, rec)
		} else {
			return records, fmt.Errorf("Unexpected record returned with CAA query to domain %s\n", domain)
		}
	}
	if len(records) == 0 {
		return records, ErrCAARecordNotFound
	}
	return records, err
}

//ParseNewRecord parses a CAA entry, and returns a new Record instance
func ParseNewRecord(caa *dns.CAA) (CAARecord, error) {
	var err error
	r := NewRecord()
	if caa.Tag == "issue" || caa.Tag == "issuewild" {
		r.Tag = caa.Tag
		fields := strings.Split(caa.Value, ";")
		if len(fields) < 1 {
			return r, fmt.Errorf("Invalid CAA record value: %s", caa.Value)
		} else {
			r.Issuer = fields[0]
		}
		for _, f := range fields[1:] {
			var key, value string
			key, value, err = parseCAAField(f)
			if strings.ToLower(key) == "validationmethods" {
				r.ValidationMethods = parseCAAValidationMethods(value)
			}
			if strings.ToLower(key) == "accounturi" {
				r.AccountUri = value
			}
		}
		r.Data = caa.String()
	}
	return r, err
}

//parseCAAField returns key-value pair of CAA attribute
func parseCAAField(input string) (string, string, error) {
	fields := strings.Split(input, "=")
	if len(fields) < 2 {
		return input, "", fmt.Errorf("Could not parse CAA field: %s", input)
	}
	return strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1]), nil
}

//parseCAAValidationMethods returns a list of valid CAA validation methods
func parseCAAValidationMethods(input string) []string {
	return strings.Split(input, ",")
}