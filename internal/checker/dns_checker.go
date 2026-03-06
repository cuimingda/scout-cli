package checker

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type DNSLookupFunc func(host, resolver string, timeout time.Duration) ([]string, error)
type SystemDNSFunc func() ([]string, error)

type DNSCheckerOptions struct {
	ExtraResolvers []string
	Lookup         DNSLookupFunc
	SystemDNS      SystemDNSFunc
	Timeout        time.Duration
}

type DNSChecker struct {
	BaseChecker
	extraResolvers []string
	lookup         DNSLookupFunc
	systemDNS      SystemDNSFunc
	timeout        time.Duration
}

type dnsCheckPlan struct {
	host          string
	resolverLabel string
	resolverAddr  string
}

type dnsResolver struct {
	label string
	addr  string
}

func NewDNSChecker(opts DNSCheckerOptions) *DNSChecker {
	lookup := opts.Lookup
	if lookup == nil {
		lookup = defaultDNSLookup
	}
	systemDNS := opts.SystemDNS
	if systemDNS == nil {
		systemDNS = currentSystemDNSes
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 3 * time.Second
	}

	return &DNSChecker{
		BaseChecker:    BaseChecker{Name: "DNS解析"},
		extraResolvers: append([]string(nil), opts.ExtraResolvers...),
		lookup:         lookup,
		systemDNS:      systemDNS,
		timeout:        timeout,
	}
}

func (c *DNSChecker) Definition() BaseChecker {
	return c.BaseChecker
}

func (c *DNSChecker) SystemDNSes() ([]string, error) {
	return c.systemDNS()
}

func (c *DNSChecker) Check(target Target) (Target, []Result) {
	if target.URL == nil {
		return target, []Result{
			failureResult(c.Name, "missing parsed URL"),
		}
	}

	plans := c.buildPlans(target)
	results := make([]Result, 0, len(plans))
	for _, plan := range plans {
		addrs, err := c.lookup(plan.host, plan.resolverAddr, c.timeout)
		if err != nil {
			results = append(results, failureResult(c.Name, fmt.Sprintf("%s在%s解析失败（%v）", plan.host, plan.resolverLabel, err)))
			continue
		}
		if len(addrs) == 0 {
			results = append(results, failureResult(c.Name, fmt.Sprintf("%s在%s解析失败（no address resolved）", plan.host, plan.resolverLabel)))
			continue
		}
		results = append(results, successResult(c.Name, fmt.Sprintf("%s在%s解析到%s", plan.host, plan.resolverLabel, strings.Join(addrs, ","))))
	}
	return target, results
}

func (c *DNSChecker) buildPlans(target Target) []dnsCheckPlan {
	host := target.URL.Hostname()
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

func currentSystemDNSes() ([]string, error) {
	f, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dnses := make([]string, 0, 3)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "nameserver" {
			continue
		}
		dnses = append(dnses, fields[1])
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(dnses) == 0 {
		return nil, fmt.Errorf("no system dns found")
	}
	return dnses, nil
}
