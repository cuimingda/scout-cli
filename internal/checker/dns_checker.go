package checker

import (
	"fmt"
	"strings"
	"time"
)

type DNSChecker struct {
	BaseChecker
	extraResolvers []string
	lookup         DNSLookupFunc
	timeout        time.Duration
}

func NewDNSChecker(opts DNSCheckerOptions) *DNSChecker {
	lookup := opts.Lookup
	if lookup == nil {
		lookup = defaultDNSLookup
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 3 * time.Second
	}

	return &DNSChecker{
		BaseChecker:    BaseChecker{Name: "DNS解析"},
		extraResolvers: append([]string(nil), opts.ExtraResolvers...),
		lookup:         lookup,
		timeout:        timeout,
	}
}

func (c *DNSChecker) Definition() BaseChecker {
	return c.BaseChecker
}

func (c *DNSChecker) Check(target URL) (URL, []Result) {
	if target.Parsed == nil {
		return target, []Result{
			failureResult(c.Name, "missing parsed URL"),
		}
	}

	plans := c.buildPlans(target)
	results := make([]Result, 0, len(plans))
	for _, plan := range plans {
		addrs, err := c.lookup(plan.host, plan.resolverAddr, c.timeout)
		if err != nil {
			results = append(results, failureResult(c.Name, fmt.Sprintf("%s在%s解析失败（%v）", plan.host, plan.resolverLabel, err)))
			continue
		}
		if len(addrs) == 0 {
			results = append(results, failureResult(c.Name, fmt.Sprintf("%s在%s解析失败（no address resolved）", plan.host, plan.resolverLabel)))
			continue
		}
		if target.ResolvedAddresses == nil {
			target.ResolvedAddresses = make(map[string][]string)
		}
		target.ResolvedAddresses[plan.resolverLabel] = append([]string(nil), addrs...)
		results = append(results, successResult(c.Name, fmt.Sprintf("%s在%s解析到%s", plan.host, plan.resolverLabel, strings.Join(addrs, ","))))
	}
	return target, results
}
