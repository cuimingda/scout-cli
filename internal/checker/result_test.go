package checker

import "testing"

func TestResultHelpers(t *testing.T) {
	success := successResult("格式检查", "ok")
	if !success.OK {
		t.Fatalf("success.OK = %v, want true", success.OK)
	}

	failure := failureResult("格式检查", "failed")
	if failure.OK {
		t.Fatalf("failure.OK = %v, want false", failure.OK)
	}
}
