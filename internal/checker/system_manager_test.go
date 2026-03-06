package checker

import "testing"

func TestSystemManager(t *testing.T) {
	dnsCalls := 0
	dnsChecker := NewSystemDNSChecker(SystemDNSCheckerOptions{
		SystemDNS: func() ([]string, error) {
			dnsCalls++
			return []string{"8.8.8.8"}, nil
		},
	})

	manager := NewSystemManager([]SystemChecker{dnsChecker})
	systemInfo, results := manager.Run()
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if dnsCalls != 1 {
		t.Fatalf("dnsCalls = %d, want 1", dnsCalls)
	}
	if len(systemInfo.DNS) != 1 || systemInfo.DNS[0] != "8.8.8.8" {
		t.Fatalf("unexpected systemInfo: %+v", systemInfo)
	}
	if results[0].Detail != "8.8.8.8" {
		t.Fatalf("unexpected result: %+v", results[0])
	}
}
