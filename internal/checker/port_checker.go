package checker

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type DialFunc func(network, address string, timeout time.Duration) error

type PortCheckerOptions struct {
	Dial    DialFunc
	Timeout time.Duration
}

type PortChecker struct {
	BaseChecker
	dial              DialFunc
	timeout           time.Duration
	defaultPortByScheme map[string]int
	portCheckByScheme map[string]bool
}

type portCheckPlan struct {
	network string
	host    string
	port    int
}

func NewPortChecker(opts PortCheckerOptions) *PortChecker {
	dial := opts.Dial
	if dial == nil {
		dial = defaultDial
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 3 * time.Second
	}

	return &PortChecker{
		BaseChecker: BaseChecker{Name: "端口检测"},
		dial:        dial,
		timeout:     timeout,
		defaultPortByScheme: map[string]int{
			"http":  80,
			"https": 443,
		},
		portCheckByScheme: map[string]bool{
			"http":  true,
			"https": true,
			"udp":   true,
		},
	}
}

func (c *PortChecker) Definition() BaseChecker {
	return c.BaseChecker
}

func (c *PortChecker) Check(target Target) (Target, []Result) {
	plan, ok, err := c.buildPlan(target)
	if err != nil {
		return target, []Result{
			failureResult(c.Name, err.Error()),
		}
	}
	if !ok {
		return target, []Result{
			successResult(c.Name, "未配置检测方案"),
		}
	}

	address := net.JoinHostPort(plan.host, strconv.Itoa(plan.port))
	if err := c.dial(plan.network, address, c.timeout); err != nil {
		return target, []Result{
			failureResult(c.Name, fmt.Sprintf("%s的%d端口未开放（%v）", plan.host, plan.port, err)),
		}
	}

	return target, []Result{
		successResult(c.Name, fmt.Sprintf("%s的%d端口开放", plan.host, plan.port)),
	}
}

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
