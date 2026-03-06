package checker

type SystemChecker interface {
	Definition() BaseChecker
	Check(System) (System, []Result)
}

type SystemManager struct {
	checkers []SystemChecker
}

func NewSystemManager(checkers []SystemChecker) SystemManager {
	return SystemManager{
		checkers: cloneSystemCheckers(checkers),
	}
}

func (m SystemManager) Run() (System, []Result) {
	target := System{}
	results := make([]Result, 0, len(m.checkers))
	for _, checker := range m.checkers {
		nextTarget, checkerResults := checker.Check(target)
		target = nextTarget
		results = append(results, checkerResults...)
	}
	return target, results
}
