package checker

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestPortChecker(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dials := []string{}
		checker := NewPortChecker(PortCheckerOptions{
			Timeout: time.Second,
			Dial: func(network, address string, _ time.Duration) error {
				dials = append(dials, fmt.Sprintf("%s://%s", network, address))
				return nil
			},
		})

		urlInfo, results := checker.Check(mustParseURL(t, "udp://tracker.opentrackr.org:1337/announce"))
		if len(results) != 1 {
			t.Fatalf("got %d results, want 1", len(results))
		}
		if !results[0].OK {
			t.Fatalf("expected success, got %+v", results[0])
		}
		if urlInfo.PortNetwork != "udp" || urlInfo.PortNumber != 1337 {
			t.Fatalf("unexpected url info: %+v", urlInfo)
		}
		if len(dials) != 1 || dials[0] != "udp://tracker.opentrackr.org:1337" {
			t.Fatalf("unexpected dials: %v", dials)
		}
	})

	t.Run("failure", func(t *testing.T) {
		checker := NewPortChecker(PortCheckerOptions{
			Timeout: time.Second,
			Dial: func(network, address string, _ time.Duration) error {
				return fmt.Errorf("mocked failure for %s://%s", network, address)
			},
		})

		_, results := checker.Check(mustParseURL(t, "https://www.google.com/"))
		if len(results) != 1 {
			t.Fatalf("got %d results, want 1", len(results))
		}
		if results[0].OK {
			t.Fatalf("expected failure, got %+v", results[0])
		}
		if !strings.Contains(results[0].Detail, "端口未开放") {
			t.Fatalf("unexpected detail: %s", results[0].Detail)
		}
	})
}
