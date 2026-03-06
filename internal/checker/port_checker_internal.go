package checker

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (c *PortChecker) buildPlan(target Target) (portCheckPlan, bool, error) {
	if target.URL == nil {
		return portCheckPlan{}, false, fmt.Errorf("missing parsed URL")
	}

	u := target.URL
	if !c.urlHasPortCheck(u) {
		return portCheckPlan{}, false, nil
	}

	port, err := c.resolvePort(u)
	if err != nil {
		return portCheckPlan{}, false, fmt.Errorf("invalid URL %q: %w", target.Raw, err)
	}
	if port == 0 {
		return portCheckPlan{}, false, nil
	}

	return portCheckPlan{
		network: c.resolveNetwork(u.Scheme),
		host:    u.Hostname(),
		port:    port,
	}, true, nil
}

func (c *PortChecker) urlHasPortCheck(u *url.URL) bool {
	scheme := strings.ToLower(u.Scheme)
	if c.hasExplicitPort(u) {
		return c.portCheckByScheme[scheme]
	}
	_, ok := c.defaultPortByScheme[scheme]
	return ok
}

func (c *PortChecker) hasExplicitPort(u *url.URL) bool {
	return u.Port() != ""
}

func (c *PortChecker) resolvePort(u *url.URL) (int, error) {
	if c.hasExplicitPort(u) {
		port, err := strconv.Atoi(u.Port())
		if err != nil {
			return 0, fmt.Errorf("invalid port %q", u.Port())
		}
		return port, nil
	}
	if port, ok := c.defaultPortByScheme[strings.ToLower(u.Scheme)]; ok {
		return port, nil
	}
	return 0, nil
}

func (c *PortChecker) resolveNetwork(scheme string) string {
	if strings.EqualFold(scheme, "udp") {
		return "udp"
	}
	return "tcp"
}

func defaultDial(network, address string, timeout time.Duration) error {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}
