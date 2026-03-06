package checker

import "testing"

func TestDNSCheckerOptionsZeroValue(t *testing.T) {
	var opts DNSCheckerOptions
	if opts.Lookup != nil {
		t.Fatal("expected nil lookup function in zero value")
	}
}
