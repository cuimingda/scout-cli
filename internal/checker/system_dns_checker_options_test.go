package checker

import "testing"

func TestSystemDNSCheckerOptionsZeroValue(t *testing.T) {
	var opts SystemDNSCheckerOptions
	if opts.SystemDNS != nil {
		t.Fatal("expected nil system dns function in zero value")
	}
}
