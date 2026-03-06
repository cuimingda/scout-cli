package checker

import "testing"

func TestSystemTargetZeroValue(t *testing.T) {
	var target SystemTarget
	if target != (SystemTarget{}) {
		t.Fatalf("unexpected target: %+v", target)
	}
}
