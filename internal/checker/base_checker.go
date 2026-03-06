package checker

type BaseChecker struct {
	Name string
}

type Checker interface {
	Definition() BaseChecker
	Check(URL) (URL, []Result)
}
