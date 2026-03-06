package checker

import "testing"

func TestSystemZeroValue(t *testing.T) {
	var systemInfo System
	if len(systemInfo.DNS) != 0 {
		t.Fatalf("systemInfo.DNS = %v, want empty", systemInfo.DNS)
	}
}
