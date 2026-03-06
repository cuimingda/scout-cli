package checker

func cloneSystemCheckers(checkers []SystemChecker) []SystemChecker {
	return append([]SystemChecker(nil), checkers...)
}
