package cmd

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func Test_buildPortCheckPlans(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		wantNetwork string
		wantHost    string
		wantPort    int
	}{
		{name: "http", raw: "http://bdremux.club/announce", wantNetwork: "tcp", wantHost: "bdremux.club", wantPort: 80},
		{name: "udp", raw: "udp://tracker.opentrackr.org:1337/announce", wantNetwork: "udp", wantHost: "tracker.opentrackr.org", wantPort: 1337},
		{name: "https", raw: "https://www.google.com/", wantNetwork: "tcp", wantHost: "www.google.com", wantPort: 443},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, err := buildPortCheckPlan(mustBuildScoutTarget(t, tt.raw))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if plan == nil {
				t.Fatal("expected plan, got nil")
			}
			if plan.network != tt.wantNetwork || plan.host != tt.wantHost || plan.port != tt.wantPort {
				t.Fatalf("unexpected plan: %+v", *plan)
			}
		})
	}
}

func Test_executePortChecks(t *testing.T) {
	raws := []string{
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

	for i, raw := range raws {
		checks := executePortChecks(mustBuildScoutTarget(t, raw))
		if len(checks) != 1 {
			t.Fatalf("raw[%d]=%s got %d checks, want 1", i, raw, len(checks))
		}
		if !checks[0].ok {
			t.Fatalf("raw[%d]=%s expected success", i, raw)
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
	raws := []string{
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

	for _, raw := range raws {
		checks := executePortChecks(mustBuildScoutTarget(t, raw))
		if len(checks) != 1 {
			t.Fatalf("raw=%s got %d checks, want 1", raw, len(checks))
		}
		if checks[0].ok {
			t.Fatalf("raw=%s expected failure", raw)
		}
		if !strings.Contains(checks[0].detail, "端口未开放") {
			t.Fatalf("unexpected detail: %s", checks[0].detail)
		}
	}
}
