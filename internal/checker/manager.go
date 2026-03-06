package checker

type Manager struct {
	formatChecker    *FormatChecker
	defaultCheckers  []Checker
	protocolCheckers map[string][]Checker
}

func NewManager(formatChecker *FormatChecker, defaultCheckers []Checker, protocolCheckers map[string][]Checker) Manager {
	return Manager{
		formatChecker:    formatChecker,
		defaultCheckers:  cloneCheckers(defaultCheckers),
		protocolCheckers: cloneProtocolCheckers(protocolCheckers),
	}
}

func (m Manager) Run(raw string) (URL, []Result) {
	target, results := m.formatChecker.Check(URL{Raw: raw})
	if len(results) == 0 || !results[0].OK {
		return target, results
	}

	for _, checker := range m.CheckersFor(target) {
		nextTarget, checkerResults := checker.Check(target)
		if nextTarget.Parsed != nil {
			target = nextTarget
		}
		results = append(results, checkerResults...)
	}
	return target, results
}

func (m Manager) CheckersFor(target URL) []Checker {
	if checkers, ok := m.protocolCheckers[target.Protocol()]; ok {
		return cloneCheckers(checkers)
	}
	return cloneCheckers(m.defaultCheckers)
}
