package checker

import "fmt"

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

func (c *FormatChecker) Check(target URL) (URL, []Result) {
	parsedURL, err := parseConnectionURL(target.Raw)
	if err != nil {
		return target, []Result{
			failureResult(c.Name, fmt.Sprintf("invalid URL %q: %v", target.Raw, err)),
		}
	}

	target.Parsed = parsedURL
	return target, []Result{
		successResult(c.Name, "输入格式合法"),
	}
}
