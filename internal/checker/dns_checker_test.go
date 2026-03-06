package checker

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestDNSChecker(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		checker := NewDNSChecker(DNSCheckerOptions{
			ExtraResolvers: []string{"223.5.5.5", "8.8.8.8"},
			Timeout:        time.Second,
			Lookup: func(host, resolver string, _ time.Duration) ([]string, error) {
				return []string{"1.1.1.1", "2.2.2.2"}, nil
			},
		})

		urlInfo, results := checker.Check(mustParseURL(t, "http://bdremux.club/announce"))
		if len(results) != 3 {
			t.Fatalf("got %d results, want 3", len(results))
		}
		if len(urlInfo.ResolvedAddresses) != 3 {
			t.Fatalf("urlInfo.ResolvedAddresses = %v, want 3 resolvers", urlInfo.ResolvedAddresses)
		}
		for _, result := range results {
			if !result.OK {
				t.Fatalf("expected success, got %+v", result)
			}
			if !strings.Contains(result.Detail, "解析到1.1.1.1,2.2.2.2") {
				t.Fatalf("unexpected detail: %s", result.Detail)
			}
		}
	})

	t.Run("skips ip host", func(t *testing.T) {
		checker := NewDNSChecker(DNSCheckerOptions{
			ExtraResolvers: []string{"223.5.5.5", "8.8.8.8"},
			Timeout:        time.Second,
		})

		_, results := checker.Check(mustParseURL(t, "http://127.0.0.1/announce"))
		if len(results) != 0 {
			t.Fatalf("expected no results, got %d", len(results))
		}
	})

	t.Run("failure", func(t *testing.T) {
		checker := NewDNSChecker(DNSCheckerOptions{
			ExtraResolvers: []string{"223.5.5.5", "8.8.8.8"},
			Timeout:        time.Second,
			Lookup: func(host, resolver string, _ time.Duration) ([]string, error) {
				return nil, fmt.Errorf("simulated dns failure")
			},
		})

		_, results := checker.Check(mustParseURL(t, "http://bdremux.club/announce"))
		if len(results) != 3 {
			t.Fatalf("got %d results, want 3", len(results))
		}
		for _, result := range results {
			if result.OK {
				t.Fatalf("expected failure, got %+v", result)
			}
			if !strings.Contains(result.Detail, "解析失败") {
				t.Fatalf("unexpected detail: %s", result.Detail)
			}
		}
	})
}
