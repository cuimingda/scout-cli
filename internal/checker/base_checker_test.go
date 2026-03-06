package checker

import "testing"

func TestBaseCheckerStoresName(t *testing.T) {
	checker := BaseChecker{Name: "DNS解析"}
	if checker.Name != "DNS解析" {
		t.Fatalf("checker.Name = %q, want %q", checker.Name, "DNS解析")
	}
}
