package checker

import "strings"

type SystemDNSChecker struct {
	BaseChecker
	systemDNS SystemDNSFunc
}

func NewSystemDNSChecker(opts SystemDNSCheckerOptions) *SystemDNSChecker {
	systemDNS := opts.SystemDNS
	if systemDNS == nil {
		systemDNS = currentSystemDNSes
	}

	return &SystemDNSChecker{
		BaseChecker: BaseChecker{Name: "当前DNS"},
		systemDNS:   systemDNS,
	}
}

func (c *SystemDNSChecker) Definition() BaseChecker {
	return c.BaseChecker
}

func (c *SystemDNSChecker) Check(target SystemTarget) (SystemTarget, []Result) {
	dnses, err := c.systemDNS()
	if err != nil {
		return target, []Result{
			failureResult(c.Name, err.Error()),
		}
	}
	return target, []Result{
		successResult(c.Name, strings.Join(dnses, ", ")),
	}
}
