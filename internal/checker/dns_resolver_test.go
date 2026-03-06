package checker

import "testing"

func TestDNSResolverHoldsValues(t *testing.T) {
	resolver := dnsResolver{label: "8.8.8.8", addr: "8.8.8.8"}
	if resolver.label != "8.8.8.8" || resolver.addr != "8.8.8.8" {
		t.Fatalf("unexpected resolver: %+v", resolver)
	}
}
