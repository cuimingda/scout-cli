package cmd

import (
	"fmt"
	"net/url"
)

type scoutTarget struct {
	raw    string
	parsed *url.URL
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

func validateConnectionURL(raw string) error {
	_, err := parseConnectionURL(raw)
	return err
}

func buildScoutTarget(raw string) (scoutTarget, error) {
	u, err := parseConnectionURL(raw)
	if err != nil {
		return scoutTarget{}, fmt.Errorf("invalid URL %q: %w", raw, err)
	}
	return scoutTarget{
		raw:    raw,
		parsed: u,
	}, nil
}

func executeFormatCheck(raw string) (scoutTarget, checkPlanResult) {
	target, err := buildScoutTarget(raw)
	if err != nil {
		return scoutTarget{}, checkPlanResult{
			name:   "文件格式检查",
			ok:     false,
			detail: err.Error(),
		}
	}
	return target, checkPlanResult{
		name:   "文件格式检查",
		ok:     true,
		detail: "输入格式合法",
	}
}
