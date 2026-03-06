package checker

import "testing"

func TestPortCheckPlanHoldsValues(t *testing.T) {
	plan := portCheckPlan{network: "tcp", host: "example.com", port: 443}
	if plan.network != "tcp" || plan.host != "example.com" || plan.port != 443 {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}
