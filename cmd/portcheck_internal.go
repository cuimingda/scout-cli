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
	url     string
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

func buildPortCheckPlans(rawURLs []string) ([]portCheckPlan, []error) {
	var plans []portCheckPlan
	var errs []error

	for _, raw := range rawURLs {
		u, err := url.Parse(raw)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid URL %q: %w", raw, err))
			continue
		}

		if !urlHasPortCheck(u) {
			continue
		}

		port, err := resolvePort(u)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid URL %q: %w", raw, err))
			continue
		}
		if port == 0 {
			continue
		}

		plans = append(plans, portCheckPlan{
			url:     raw,
			network: resolveNetwork(u.Scheme),
			host:    u.Hostname(),
			port:    port,
		})
	}

	return plans, errs
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
		return portCheckError(plan.url, plan, err)
	}
	return nil
}

func portCheckError(url string, plan portCheckPlan, err error) error {
	return fmt.Errorf("port check failed for %q (%s:%d via %s): %w", url, plan.host, plan.port, plan.network, err)
}
