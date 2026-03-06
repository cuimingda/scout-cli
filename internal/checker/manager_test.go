package checker

import (
	"fmt"
	"testing"
	"time"
)

func TestManager(t *testing.T) {
	formatChecker := NewFormatChecker()

	portCalls := 0
	portChecker := NewPortChecker(PortCheckerOptions{
		Timeout: time.Second,
		Dial: func(network, address string, _ time.Duration) error {
			portCalls++
			return nil
		},
	})

	dnsCalls := 0
	dnsChecker := NewDNSChecker(DNSCheckerOptions{
		ExtraResolvers: []string{"223.5.5.5", "8.8.8.8"},
		Timeout:        time.Second,
		SystemDNS: func() ([]string, error) {
			return []string{"8.8.8.8"}, nil
		},
		Lookup: func(host, resolver string, _ time.Duration) ([]string, error) {
			dnsCalls++
			return []string{"1.1.1.1"}, nil
		},
	})

	manager := NewManager(
		formatChecker,
		dnsChecker,
		[]Checker{dnsChecker},
		map[string][]Checker{
			"http":  []Checker{portChecker, dnsChecker},
			"https": []Checker{portChecker, dnsChecker},
			"udp":   []Checker{portChecker, dnsChecker},
		},
	)

	_, results := manager.Run("http://example.com")
	if len(results) != 5 {
		t.Fatalf("got %d results, want 5", len(results))
	}
	if portCalls != 1 || dnsCalls != 3 {
		t.Fatalf("unexpected checker calls: port=%d dns=%d", portCalls, dnsCalls)
	}

	portCalls = 0
	dnsCalls = 0

	_, results = manager.Run("ftp://example.com/resource")
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4", len(results))
	}
	if portCalls != 0 || dnsCalls != 3 {
		t.Fatalf("unexpected checker calls: port=%d dns=%d", portCalls, dnsCalls)
	}

	portCalls = 0
	dnsCalls = 0

	_, results = manager.Run("google.com")
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].OK {
		t.Fatalf("expected format failure, got %+v", results[0])
	}
	if portCalls != 0 || dnsCalls != 0 {
		t.Fatalf("unexpected checker calls after format failure: port=%d dns=%d", portCalls, dnsCalls)
	}

	systemDNS, err := manager.SystemDNSes()
	if err != nil {
		t.Fatalf("SystemDNSes() error = %v", err)
	}
	if len(systemDNS) != 1 || systemDNS[0] != "8.8.8.8" {
		t.Fatalf("systemDNS = %v, want [8.8.8.8]", systemDNS)
	}
}

func TestManagerSystemDNSesRequiresDNSChecker(t *testing.T) {
	manager := NewManager(NewFormatChecker(), nil, nil, nil)
	if _, err := manager.SystemDNSes(); err == nil {
		t.Fatal("expected error when dns checker is not configured")
	} else if err.Error() != fmt.Sprintf("%s", "dns checker not configured") {
		t.Fatalf("unexpected error: %v", err)
	}
}
