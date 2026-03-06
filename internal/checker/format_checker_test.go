package checker

import (
	"strings"
	"testing"
)

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
