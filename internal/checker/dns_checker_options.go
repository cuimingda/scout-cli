package checker

import "time"

type DNSLookupFunc func(host, resolver string, timeout time.Duration) ([]string, error)

type DNSCheckerOptions struct {
	ExtraResolvers []string
	Lookup         DNSLookupFunc
	Timeout        time.Duration
}
