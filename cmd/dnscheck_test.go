package cmd

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func Test_buildDNSCheckPlans(t *testing.T) {
	extraDNS := []string{
		"223.5.5.5",
		"8.8.8.8",
	}
	urls := []string{
		"http://bdremux.club/announce",
		"http://192.168.1.1/announce",
		"https://www.google.com/",
	}
	plans, errs := buildDNSCheckPlans(urls, extraDNS)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(plans) != 6 {
		t.Fatalf("got %d plans, want 6", len(plans))
	}

	hosts := []string{"bdremux.club", "www.google.com"}
	resolvers := append([]dnsResolver{{label: "当前DNS", addr: ""}}, dnsResolver{label: "223.5.5.5", addr: "223.5.5.5"}, dnsResolver{label: "8.8.8.8", addr: "8.8.8.8"})
	for i, expected := range hosts {
		base := i * 3
		for j, want := range resolvers {
			plan := plans[base+j]
			if plan.host != expected {
				t.Fatalf("plan[%d].host got %q, want %q", base+j, plan.host, expected)
			}
			if plan.resolverLabel != want.label {
				t.Fatalf("plan[%d].resolverLabel got %q, want %q", base+j, plan.resolverLabel, want.label)
			}
			if plan.resolverAddr != want.addr {
				t.Fatalf("plan[%d].resolverAddr got %q, want %q", base+j, plan.resolverAddr, want.addr)
			}
		}
	}
}

func Test_executeDNSChecks(t *testing.T) {
	urls := []string{
		"http://bdremux.club/announce",
		"https://www.google.com/",
	}
	origLookup := detectDNSLookup
	defer func() { detectDNSLookup = origLookup }()
	detectDNSLookup = func(host, resolver string, _ time.Duration) ([]string, error) {
		return []string{"1.1.1.1", "2.2.2.2"}, nil
	}

	reports := executeDNSChecks(urls, []string{"223.5.5.5", "8.8.8.8"})
	if len(reports) != len(urls) {
		t.Fatalf("got %d reports, want %d", len(reports), len(urls))
	}

	for i, report := range reports {
		if len(report.checks) != 3 {
			t.Fatalf("report[%d] checks = %d, want 3", i, len(report.checks))
		}
		for _, check := range report.checks {
			if !check.ok {
				t.Fatalf("report %s should all succeed in DNS checks", report.url)
			}
			if !strings.Contains(check.detail, "解析到1.1.1.1,2.2.2.2") {
				t.Fatalf("unexpected detail: %s", check.detail)
			}
		}
	}
}

func Test_executeDNSChecks_skips_ip_host(t *testing.T) {
	reports := executeDNSChecks([]string{"http://127.0.0.1/announce"}, []string{"223.5.5.5", "8.8.8.8"})
	if len(reports) != 1 {
		t.Fatalf("got %d reports, want 1", len(reports))
	}
	if len(reports[0].checks) != 0 {
		t.Fatalf("expected no DNS checks for IP host")
	}
}

func Test_executeDNSChecks_collects_all_errors(t *testing.T) {
	urls := []string{"http://bdremux.club/announce", "https://www.google.com/"}
	origLookup := detectDNSLookup
	defer func() { detectDNSLookup = origLookup }()
	detectDNSLookup = func(host, resolver string, _ time.Duration) ([]string, error) {
		return nil, fmt.Errorf("simulated dns failure")
	}

	reports := executeDNSChecks(urls, []string{"223.5.5.5", "8.8.8.8"})
	for _, report := range reports {
		if len(report.checks) != 3 {
			t.Fatalf("url=%s expected 3 checks, got %d", report.url, len(report.checks))
		}
		for _, check := range report.checks {
			if check.ok {
				t.Fatalf("url=%s expected failure", report.url)
			}
			if !strings.Contains(check.detail, "解析失败") {
				t.Fatalf("unexpected detail: %s", check.detail)
			}
		}
	}
}
