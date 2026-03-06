package checker

import (
	neturl "net/url"
	"strings"
)

type URL struct {
	Raw               string
	Parsed            *neturl.URL
	PortNetwork       string
	PortNumber        int
	ResolvedAddresses map[string][]string
}

func (u URL) Protocol() string {
	if u.Parsed == nil {
		return ""
	}
	return strings.ToLower(u.Parsed.Scheme)
}
