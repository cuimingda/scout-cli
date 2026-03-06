package checker

import "testing"

func mustParseURL(t *testing.T, raw string) URL {
	t.Helper()

	parsedURL, err := parseConnectionURL(raw)
	if err != nil {
		t.Fatalf("parseConnectionURL(%q) error = %v", raw, err)
	}
	return URL{
		Raw:    raw,
		Parsed: parsedURL,
	}
}

func TestURLProtocol(t *testing.T) {
	urlInfo := mustParseURL(t, "HTTPS://www.google.com")
	if urlInfo.Protocol() != "https" {
		t.Fatalf("urlInfo.Protocol() = %q, want %q", urlInfo.Protocol(), "https")
	}
}
