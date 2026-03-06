package cmd

import "fmt"

type checkPlanResult struct {
	name   string
	ok     bool
	detail string
}

func executePortChecks(target scoutTarget) []checkPlanResult {
	plan, err := buildPortCheckPlan(target)
	if err != nil {
		return []checkPlanResult{{
			name:   "端口检测",
			ok:     false,
			detail: err.Error(),
		}}
	}

	if plan == nil {
		return []checkPlanResult{{
			name:   "端口检测",
			ok:     true,
			detail: "未配置检测方案",
		}}
	}

	if err := executePortCheck(*plan); err != nil {
		return []checkPlanResult{{
			name:   "端口检测",
			ok:     false,
			detail: err.Error(),
		}}
	}

	return []checkPlanResult{{
		name:   "端口检测",
		ok:     true,
		detail: fmt.Sprintf("%s的%d端口开放", plan.host, plan.port),
	}}
}

func executePortChecksStreaming(target scoutTarget, write func(checkPlanResult)) {
	for _, check := range executePortChecks(target) {
		write(check)
	}
}
