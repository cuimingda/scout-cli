package checker

import (
	"fmt"
	"net/url"
)

type FormatChecker struct {
	BaseChecker
}

func NewFormatChecker() *FormatChecker {
	return &FormatChecker{
		BaseChecker: BaseChecker{Name: "文件格式检查"},
	}
}

func (c *FormatChecker) Definition() BaseChecker {
	return c.BaseChecker
}

func (c *FormatChecker) Check(target Target) (Target, []Result) {
	parsedURL, err := parseConnectionURL(target.Raw)
	if err != nil {
		return target, []Result{
			failureResult(c.Name, fmt.Sprintf("invalid URL %q: %v", target.Raw, err)),
		}
	}

	target.URL = parsedURL
	return target, []Result{
		successResult(c.Name, "输入格式合法"),
	}
}

func parseConnectionURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		return nil, fmt.Errorf("missing protocol")
	}
	if u.Host == "" {
		return nil, fmt.Errorf("missing host")
	}
	return u, nil
}
