package checker

import "testing"

func TestPortCheckerOptionsZeroValue(t *testing.T) {
	var opts PortCheckerOptions
	if opts.Dial != nil {
		t.Fatal("expected nil dial function in zero value")
	}
}
