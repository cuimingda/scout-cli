package checker

import (
	"fmt"
	"testing"
)

func TestSystemDNSChecker(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		checker := NewSystemDNSChecker(SystemDNSCheckerOptions{
			SystemDNS: func() ([]string, error) {
				return []string{"8.8.8.8", "223.5.5.5"}, nil
			},
		})

		_, results := checker.Check(SystemTarget{})
		if len(results) != 1 {
			t.Fatalf("got %d results, want 1", len(results))
		}
		if !results[0].OK {
			t.Fatalf("expected success, got %+v", results[0])
		}
		if results[0].Name != "当前DNS" || results[0].Detail != "8.8.8.8, 223.5.5.5" {
			t.Fatalf("unexpected result: %+v", results[0])
		}
	})

	t.Run("failure", func(t *testing.T) {
		checker := NewSystemDNSChecker(SystemDNSCheckerOptions{
			SystemDNS: func() ([]string, error) {
				return nil, fmt.Errorf("simulated system dns failure")
			},
		})

		_, results := checker.Check(SystemTarget{})
		if len(results) != 1 {
			t.Fatalf("got %d results, want 1", len(results))
		}
		if results[0].OK {
			t.Fatalf("expected failure, got %+v", results[0])
		}
		if results[0].Detail != "simulated system dns failure" {
			t.Fatalf("unexpected result: %+v", results[0])
		}
	})
}
