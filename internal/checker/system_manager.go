package checker

type SystemChecker interface {
	Definition() BaseChecker
	Check(SystemTarget) (SystemTarget, []Result)
}

type SystemManager struct {
	checkers []SystemChecker
}

func NewSystemManager(checkers []SystemChecker) SystemManager {
	return SystemManager{
		checkers: cloneSystemCheckers(checkers),
	}
}

func (m SystemManager) Run() (SystemTarget, []Result) {
	target := SystemTarget{}
	results := make([]Result, 0, len(m.checkers))
	for _, checker := range m.checkers {
		nextTarget, checkerResults := checker.Check(target)
		target = nextTarget
		results = append(results, checkerResults...)
	}
	return target, results
}
