package checker

import "testing"

type stubSystemChecker struct {
	BaseChecker
	calls   int
	results []Result
}

func (s *stubSystemChecker) Definition() BaseChecker {
	return s.BaseChecker
}

func (s *stubSystemChecker) Check(target SystemTarget) (SystemTarget, []Result) {
	s.calls++
	return target, s.results
}

func TestSystemManager(t *testing.T) {
	dnsChecker := &stubSystemChecker{
		BaseChecker: BaseChecker{Name: "当前DNS"},
		results:     []Result{successResult("当前DNS", "8.8.8.8")},
	}

	manager := NewSystemManager([]SystemChecker{dnsChecker})
	_, results := manager.Run()
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if dnsChecker.calls != 1 {
		t.Fatalf("dnsChecker.calls = %d, want 1", dnsChecker.calls)
	}
	if results[0].Detail != "8.8.8.8" {
		t.Fatalf("unexpected result: %+v", results[0])
	}
}
