package checker

import "fmt"

type Manager struct {
	formatChecker    *FormatChecker
	dnsChecker       *DNSChecker
	defaultCheckers  []Checker
	protocolCheckers map[string][]Checker
}

func NewManager(formatChecker *FormatChecker, dnsChecker *DNSChecker, defaultCheckers []Checker, protocolCheckers map[string][]Checker) Manager {
	return Manager{
		formatChecker:    formatChecker,
		dnsChecker:       dnsChecker,
		defaultCheckers:  cloneCheckers(defaultCheckers),
		protocolCheckers: cloneProtocolCheckers(protocolCheckers),
	}
}

func (m Manager) Run(raw string) (Target, []Result) {
	target, results := m.formatChecker.Check(Target{Raw: raw})
	if len(results) == 0 || !results[0].OK {
		return target, results
	}

	for _, checker := range m.CheckersFor(target) {
		nextTarget, checkerResults := checker.Check(target)
		if nextTarget.URL != nil {
			target = nextTarget
		}
		results = append(results, checkerResults...)
	}
	return target, results
}

func (m Manager) CheckersFor(target Target) []Checker {
	if checkers, ok := m.protocolCheckers[target.Protocol()]; ok {
		return cloneCheckers(checkers)
	}
	return cloneCheckers(m.defaultCheckers)
}

func (m Manager) SystemDNSes() ([]string, error) {
	if m.dnsChecker == nil {
		return nil, fmt.Errorf("dns checker not configured")
	}
	return m.dnsChecker.SystemDNSes()
}
