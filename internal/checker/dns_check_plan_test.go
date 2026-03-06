package checker

import "testing"

func TestDNSCheckPlanHoldsValues(t *testing.T) {
	plan := dnsCheckPlan{host: "example.com", resolverLabel: "当前DNS", resolverAddr: ""}
	if plan.host != "example.com" || plan.resolverLabel != "当前DNS" {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}
