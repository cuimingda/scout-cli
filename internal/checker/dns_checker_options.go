package checker

import "time"

type DNSLookupFunc func(host, resolver string, timeout time.Duration) ([]string, error)
type SystemDNSFunc func() ([]string, error)

type DNSCheckerOptions struct {
	ExtraResolvers []string
	Lookup         DNSLookupFunc
	SystemDNS      SystemDNSFunc
	Timeout        time.Duration
}
