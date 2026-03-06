package cmd

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type portCheckPlan struct {
	network string
	host    string
	port    int
}

var defaultPortByScheme = map[string]int{
	"http":  80,
	"https": 443,
}

var portCheckByScheme = map[string]bool{
	"http":  true,
	"https": true,
	"udp":  true,
}

var detectDial = func(network, address string, timeout time.Duration) error {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}

func buildPortCheckPlan(target scoutTarget) (*portCheckPlan, error) {
	u := target.parsed
	if !urlHasPortCheck(u) {
		return nil, nil
	}

	port, err := resolvePort(u)
	if err != nil {
		return nil, fmt.Errorf("invalid URL %q: %w", target.raw, err)
	}
	if port == 0 {
		return nil, nil
	}

	return &portCheckPlan{
		network: resolveNetwork(u.Scheme),
		host:    u.Hostname(),
		port:    port,
	}, nil
}

func urlHasPortCheck(u *url.URL) bool {
	scheme := strings.ToLower(u.Scheme)
	if hasExplicitPort(u) {
		return portCheckByScheme[scheme]
	}
	_, ok := defaultPortByScheme[scheme]
	return ok
}

func hasExplicitPort(u *url.URL) bool {
	return u.Port() != ""
}

func resolvePort(u *url.URL) (int, error) {
	if hasExplicitPort(u) {
		port, err := strconv.Atoi(u.Port())
		if err != nil {
			return 0, fmt.Errorf("invalid port %q", u.Port())
		}
		return port, nil
	}
	if port, ok := defaultPortByScheme[strings.ToLower(u.Scheme)]; ok {
		return port, nil
	}
	return 0, nil
}

func resolveNetwork(scheme string) string {
	if strings.EqualFold(scheme, "udp") {
		return "udp"
	}
	return "tcp"
}

func executePortCheck(plan portCheckPlan) error {
	address := net.JoinHostPort(plan.host, strconv.Itoa(plan.port))
	if err := detectDial(plan.network, address, 3*time.Second); err != nil {
		return portCheckError(plan, err)
	}
	return nil
}

func portCheckError(plan portCheckPlan, err error) error {
	return fmt.Errorf("%s的%d端口未开放（%w）", plan.host, plan.port, err)
}
