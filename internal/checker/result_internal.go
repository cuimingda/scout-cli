package checker

func successResult(name, detail string) Result {
	return Result{
		Name:   name,
		OK:     true,
		Detail: detail,
	}
}

func failureResult(name, detail string) Result {
	return Result{
		Name:   name,
		OK:     false,
		Detail: detail,
	}
}
