package checker

import (
	"context"
	"net"
	"time"
)

func (c *DNSChecker) buildPlans(target URL) []dnsCheckPlan {
	host := target.Parsed.Hostname()
	if host == "" {
		return nil
	}
	if ip := net.ParseIP(host); ip != nil {
		return nil
	}

	resolvers := make([]dnsResolver, 0, len(c.extraResolvers)+1)
	resolvers = append(resolvers, dnsResolver{label: "当前DNS", addr: ""})
	for _, dnsAddr := range c.extraResolvers {
		resolvers = append(resolvers, dnsResolver{label: dnsAddr, addr: dnsAddr})
	}

	plans := make([]dnsCheckPlan, 0, len(resolvers))
	for _, resolver := range resolvers {
		plans = append(plans, dnsCheckPlan{
			host:          host,
			resolverLabel: resolver.label,
			resolverAddr:  resolver.addr,
		})
	}
	return plans
}

func defaultDNSLookup(host, resolver string, timeout time.Duration) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if resolver == "" {
		return net.DefaultResolver.LookupHost(ctx, host)
	}

	resolverClient := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, net.JoinHostPort(resolver, "53"))
		},
	}
	return resolverClient.LookupHost(ctx, host)
}
