package checker

import (
	"net/url"
	"strings"
)

type Result struct {
	Name   string
	OK     bool
	Detail string
}

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

type BaseChecker struct {
	Name string
}

type Checker interface {
	Definition() BaseChecker
	Check(Target) (Target, []Result)
}

func successResult(name, detail string) Result {
	return Result{
		Name:   name,
		OK:     true,
		Detail: detail,
	}
}

func failureResult(name, detail string) Result {
	return Result{
		Name:   name,
		OK:     false,
		Detail: detail,
	}
}
