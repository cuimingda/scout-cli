package cmd

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"net/url"
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

var dnsResolvers = []dnsResolver{
	{label: "当前DNS", addr: ""},
	{label: "8.8.8.8", addr: "8.8.8.8"},
	{label: "223.5.5.5", addr: "223.5.5.5"},
}

var detectSystemDNS = func() ([]string, error) {
	return currentSystemDNSes()
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

func buildDNSCheckPlans(rawURLs []string) ([]dnsCheckPlan, []error) {
	var plans []dnsCheckPlan
	var errs []error

	for _, raw := range rawURLs {
		u, err := url.Parse(raw)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid URL %q: %w", raw, err))
			continue
		}
		if u.Host == "" {
			errs = append(errs, fmt.Errorf("invalid URL %q: missing host", raw))
			continue
		}

		host := u.Hostname()
		if host == "" {
			errs = append(errs, fmt.Errorf("invalid URL %q: missing host", raw))
			continue
		}
		if ip := net.ParseIP(host); ip != nil {
			continue
		}

		for _, resolver := range dnsResolvers {
			plans = append(plans, dnsCheckPlan{
				url:           raw,
				host:          host,
				resolverLabel: resolver.label,
				resolverAddr:  resolver.addr,
			})
		}
	}
	return plans, errs
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
