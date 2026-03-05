package cmd

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func Test_buildPortCheckPlans(t *testing.T) {
	urls := []string{
		"http://bdremux.club/announce",
		"udp://tracker.opentrackr.org:1337/announce",
		"https://www.google.com/",
	}
	plans, errs := buildPortCheckPlans(urls)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(plans) != 3 {
		t.Fatalf("got %d plans, want 3", len(plans))
	}

	if plans[0].url != "http://bdremux.club/announce" {
		t.Fatalf("got url %q", plans[0].url)
	}
	if plans[0].network != "tcp" || plans[0].host != "bdremux.club" || plans[0].port != 80 {
		t.Fatalf("unexpected http plan: %+v", plans[0])
	}
	if plans[1].network != "udp" || plans[1].host != "tracker.opentrackr.org" || plans[1].port != 1337 {
		t.Fatalf("unexpected udp plan: %+v", plans[1])
	}
	if plans[2].network != "tcp" || plans[2].host != "www.google.com" || plans[2].port != 443 {
		t.Fatalf("unexpected https plan: %+v", plans[2])
	}
}

func Test_executePortChecks(t *testing.T) {
	urls := []string{
		"http://bdremux.club/announce",
		"udp://tracker.opentrackr.org:1337/announce",
		"https://www.google.com/",
	}
	dials := []string{}
	origDial := detectDial
	defer func() {
		detectDial = origDial
	}()
	detectDial = func(network, address string, _ time.Duration) error {
		dials = append(dials, fmt.Sprintf("%s://%s", network, address))
		return nil
	}

	reports := executePortChecks(urls)
	if len(reports) != len(urls) {
		t.Fatalf("got %d reports, want %d", len(reports), len(urls))
	}

	for i, report := range reports {
		if len(report.checks) != 1 {
			t.Fatalf("url[%d]=%s got %d checks, want 1", i, report.url, len(report.checks))
		}
		if !report.checks[0].ok {
			t.Fatalf("url[%d]=%s expected success", i, report.url)
		}
	}

	if len(dials) != 3 {
		t.Fatalf("got %d dials, want 3", len(dials))
	}
	seen := strings.Join(dials, ",")
	if !strings.Contains(seen, "tcp://bdremux.club:80") ||
		!strings.Contains(seen, "udp://tracker.opentrackr.org:1337") ||
		!strings.Contains(seen, "tcp://www.google.com:443") {
		t.Fatalf("unexpected dial targets: %v", dials)
	}
}

func Test_executePortChecks_collects_all_errors(t *testing.T) {
	urls := []string{
		"udp://tracker.opentrackr.org:1337/announce",
		"https://www.google.com/",
	}

	origDial := detectDial
	defer func() {
		detectDial = origDial
	}()
	detectDial = func(network, address string, _ time.Duration) error {
		return fmt.Errorf("mocked failure for %s://%s", network, address)
	}

	reports := executePortChecks(urls)
	for _, report := range reports {
		if len(report.checks) != 1 {
			t.Fatalf("url=%s got %d checks, want 1", report.url, len(report.checks))
		}
		if report.checks[0].ok {
			t.Fatalf("url=%s expected failure", report.url)
		}
		if !strings.Contains(report.checks[0].detail, "port check failed") {
			t.Fatalf("unexpected detail: %s", report.checks[0].detail)
		}
	}
}
