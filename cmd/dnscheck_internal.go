package cmd

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type dnsCheckPlan struct {
	url           string
	host          string
	resolverLabel string
	resolverAddr  string
}

type dnsResolver struct {
	label string
	addr  string
}

var detectDNSLookup = func(host, resolver string, _ time.Duration) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
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

var detectSystemDNS = func() ([]string, error) {
	return currentSystemDNSes()
}

func buildDNSCheckPlans(target scoutTarget, extraDNS []string) []dnsCheckPlan {
	resolvers := make([]dnsResolver, 0, len(extraDNS)+1)
	resolvers = append(resolvers, dnsResolver{label: "当前DNS", addr: ""})
	for _, dnsAddr := range extraDNS {
		resolvers = append(resolvers, dnsResolver{label: dnsAddr, addr: dnsAddr})
	}

	host := target.parsed.Hostname()
	if host == "" {
		return nil
	}
	if ip := net.ParseIP(host); ip != nil {
		return nil
	}

	plans := make([]dnsCheckPlan, 0, len(resolvers))
	for _, resolver := range resolvers {
		plans = append(plans, dnsCheckPlan{
			url:           target.raw,
			host:          host,
			resolverLabel: resolver.label,
			resolverAddr:  resolver.addr,
		})
	}
	return plans
}

func executeDNSChecksWithResolvers(target scoutTarget, extraDNS []string) []checkPlanResult {
	plans := buildDNSCheckPlans(target, extraDNS)
	checks := make([]checkPlanResult, 0, len(plans))
	for _, plan := range plans {
		ips, err := executeDNSCheck(plan)
		if err != nil {
			checks = append(checks, checkPlanResult{
				name:   "DNS解析",
				ok:     false,
				detail: err.Error(),
			})
			continue
		}
		checks = append(checks, checkPlanResult{
			name:   "DNS解析",
			ok:     true,
			detail: fmt.Sprintf("%s在%s解析到%s", plan.host, plan.resolverLabel, strings.Join(ips, ",")),
		})
	}
	return checks
}

func executeDNSChecksStreamingWithResolvers(target scoutTarget, extraDNS []string, write func(checkPlanResult)) {
	for _, check := range executeDNSChecksWithResolvers(target, extraDNS) {
		write(check)
	}
}

func executeDNSCheck(plan dnsCheckPlan) ([]string, error) {
	addrs, err := detectDNSLookup(plan.host, plan.resolverAddr, 3*time.Second)
	if err != nil {
		return nil, dnsCheckError(plan, err)
	}
	if len(addrs) == 0 {
		return nil, dnsCheckError(plan, fmt.Errorf("no address resolved"))
	}
	return addrs, nil
}

func dnsCheckError(plan dnsCheckPlan, err error) error {
	return fmt.Errorf("%s在%s解析失败（%w）", plan.host, plan.resolverLabel, err)
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
