package checker

import "testing"

func TestTargetProtocol(t *testing.T) {
	target := mustParseTarget(t, "HTTPS://www.google.com")
	if target.Protocol() != "https" {
		t.Fatalf("target.Protocol() = %q, want %q", target.Protocol(), "https")
	}
}
