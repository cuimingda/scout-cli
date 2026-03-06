package checker

func cloneCheckers(checkers []Checker) []Checker {
	return append([]Checker(nil), checkers...)
}

func cloneProtocolCheckers(protocolCheckers map[string][]Checker) map[string][]Checker {
	cloned := make(map[string][]Checker, len(protocolCheckers))
	for protocol, checkers := range protocolCheckers {
		cloned[protocol] = cloneCheckers(checkers)
	}
	return cloned
}
