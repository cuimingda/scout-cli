package checker

import (
	"net/url"
	"strings"
)

type Target struct {
	Raw string
	URL *url.URL
}

func (t Target) Protocol() string {
	if t.URL == nil {
		return ""
	}
	return strings.ToLower(t.URL.Scheme)
}
