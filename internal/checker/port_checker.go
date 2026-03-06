package checker

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

type PortChecker struct {
	BaseChecker
	dial                DialFunc
	timeout             time.Duration
	defaultPortByScheme map[string]int
	portCheckByScheme   map[string]bool
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

func (c *PortChecker) Check(target URL) (URL, []Result) {
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

	target.PortNetwork = plan.network
	target.PortNumber = plan.port
	return target, []Result{
		successResult(c.Name, fmt.Sprintf("%s的%d端口开放", plan.host, plan.port)),
	}
}
