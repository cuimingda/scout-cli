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
	plans := buildDNSCheckPlans(mustBuildScoutTarget(t, "http://bdremux.club/announce"), extraDNS)
	if len(plans) != 3 {
		t.Fatalf("got %d plans, want 3", len(plans))
	}

	resolvers := []dnsResolver{
		{label: "当前DNS", addr: ""},
		{label: "223.5.5.5", addr: "223.5.5.5"},
		{label: "8.8.8.8", addr: "8.8.8.8"},
	}
	for i, want := range resolvers {
		plan := plans[i]
		if plan.host != "bdremux.club" {
			t.Fatalf("plan[%d].host got %q, want %q", i, plan.host, "bdremux.club")
		}
		if plan.resolverLabel != want.label {
			t.Fatalf("plan[%d].resolverLabel got %q, want %q", i, plan.resolverLabel, want.label)
		}
		if plan.resolverAddr != want.addr {
			t.Fatalf("plan[%d].resolverAddr got %q, want %q", i, plan.resolverAddr, want.addr)
		}
	}

	if got := len(buildDNSCheckPlans(mustBuildScoutTarget(t, "http://192.168.1.1/announce"), extraDNS)); got != 0 {
		t.Fatalf("ip host should not create dns plans, got %d", got)
	}
}

func Test_executeDNSChecks(t *testing.T) {
	origLookup := detectDNSLookup
	defer func() { detectDNSLookup = origLookup }()
	detectDNSLookup = func(host, resolver string, _ time.Duration) ([]string, error) {
		return []string{"1.1.1.1", "2.2.2.2"}, nil
	}

	checks := executeDNSChecks(mustBuildScoutTarget(t, "http://bdremux.club/announce"), []string{"223.5.5.5", "8.8.8.8"})
	if len(checks) != 3 {
		t.Fatalf("got %d checks, want 3", len(checks))
	}

	for _, check := range checks {
		if !check.ok {
			t.Fatal("expected all dns checks to succeed")
		}
		if !strings.Contains(check.detail, "解析到1.1.1.1,2.2.2.2") {
			t.Fatalf("unexpected detail: %s", check.detail)
		}
	}
}

func Test_executeDNSChecks_skips_ip_host(t *testing.T) {
	checks := executeDNSChecks(mustBuildScoutTarget(t, "http://127.0.0.1/announce"), []string{"223.5.5.5", "8.8.8.8"})
	if len(checks) != 0 {
		t.Fatalf("expected no DNS checks for IP host")
	}
}

func Test_executeDNSChecks_collects_all_errors(t *testing.T) {
	origLookup := detectDNSLookup
	defer func() { detectDNSLookup = origLookup }()
	detectDNSLookup = func(host, resolver string, _ time.Duration) ([]string, error) {
		return nil, fmt.Errorf("simulated dns failure")
	}

	checks := executeDNSChecks(mustBuildScoutTarget(t, "http://bdremux.club/announce"), []string{"223.5.5.5", "8.8.8.8"})
	if len(checks) != 3 {
		t.Fatalf("expected 3 checks, got %d", len(checks))
	}
	for _, check := range checks {
		if check.ok {
			t.Fatal("expected failure")
		}
		if !strings.Contains(check.detail, "解析失败") {
			t.Fatalf("unexpected detail: %s", check.detail)
		}
	}
}
