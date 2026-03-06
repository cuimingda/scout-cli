package checker

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

type stubChecker struct {
	BaseChecker
	calls   int
	results []Result
}

func (s *stubChecker) Definition() BaseChecker {
	return s.BaseChecker
}

func (s *stubChecker) Check(target Target) (Target, []Result) {
	s.calls++
	return target, s.results
}

func mustParseTarget(t *testing.T, raw string) Target {
	t.Helper()

	parsedURL, err := parseConnectionURL(raw)
	if err != nil {
		t.Fatalf("parseConnectionURL(%q) error = %v", raw, err)
	}
	return Target{
		Raw: raw,
		URL: parsedURL,
	}
}

func TestFormatChecker(t *testing.T) {
	checker := NewFormatChecker()

	target, results := checker.Check(Target{Raw: "https://www.google.com/sitemap.xml"})
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if !results[0].OK {
		t.Fatalf("expected success, got %+v", results[0])
	}
	if target.URL == nil {
		t.Fatal("expected parsed url")
	}

	_, results = checker.Check(Target{Raw: "google.com"})
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].OK {
		t.Fatalf("expected failure, got %+v", results[0])
	}
	if !strings.Contains(results[0].Detail, "missing protocol") {
		t.Fatalf("unexpected detail: %s", results[0].Detail)
	}
}

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

		_, results := checker.Check(mustParseTarget(t, "udp://tracker.opentrackr.org:1337/announce"))
		if len(results) != 1 {
			t.Fatalf("got %d results, want 1", len(results))
		}
		if !results[0].OK {
			t.Fatalf("expected success, got %+v", results[0])
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

		_, results := checker.Check(mustParseTarget(t, "https://www.google.com/"))
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

func TestDNSChecker(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		checker := NewDNSChecker(DNSCheckerOptions{
			ExtraResolvers: []string{"223.5.5.5", "8.8.8.8"},
			Timeout:        time.Second,
			Lookup: func(host, resolver string, _ time.Duration) ([]string, error) {
				return []string{"1.1.1.1", "2.2.2.2"}, nil
			},
		})

		_, results := checker.Check(mustParseTarget(t, "http://bdremux.club/announce"))
		if len(results) != 3 {
			t.Fatalf("got %d results, want 3", len(results))
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

		_, results := checker.Check(mustParseTarget(t, "http://127.0.0.1/announce"))
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

		_, results := checker.Check(mustParseTarget(t, "http://bdremux.club/announce"))
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

func TestManager(t *testing.T) {
	formatChecker := NewFormatChecker()
	portChecker := &stubChecker{
		BaseChecker: BaseChecker{Name: "端口检测"},
		results:     []Result{successResult("端口检测", "port ok")},
	}
	dnsChecker := &stubChecker{
		BaseChecker: BaseChecker{Name: "DNS解析"},
		results:     []Result{successResult("DNS解析", "dns ok")},
	}

	manager := NewManager(
		formatChecker,
		nil,
		[]Checker{dnsChecker},
		map[string][]Checker{
			"http":  []Checker{portChecker, dnsChecker},
			"https": []Checker{portChecker, dnsChecker},
			"udp":   []Checker{portChecker, dnsChecker},
		},
	)

	_, results := manager.Run("http://example.com")
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}
	if portChecker.calls != 1 || dnsChecker.calls != 1 {
		t.Fatalf("unexpected checker calls: port=%d dns=%d", portChecker.calls, dnsChecker.calls)
	}

	portChecker.calls = 0
	dnsChecker.calls = 0

	_, results = manager.Run("ftp://example.com/resource")
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if portChecker.calls != 0 || dnsChecker.calls != 1 {
		t.Fatalf("unexpected checker calls: port=%d dns=%d", portChecker.calls, dnsChecker.calls)
	}

	portChecker.calls = 0
	dnsChecker.calls = 0

	_, results = manager.Run("google.com")
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].OK {
		t.Fatalf("expected format failure, got %+v", results[0])
	}
	if portChecker.calls != 0 || dnsChecker.calls != 0 {
		t.Fatalf("unexpected checker calls after format failure: port=%d dns=%d", portChecker.calls, dnsChecker.calls)
	}
}
