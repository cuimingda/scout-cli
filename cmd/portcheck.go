package cmd

type checkPlanResult struct {
	name   string
	ok     bool
	detail string
}

type urlCheckReport struct {
	url    string
	checks []checkPlanResult
}

func executePortChecks(rawURLs []string) []urlCheckReport {
	reports := make([]urlCheckReport, 0, len(rawURLs))
	for _, raw := range rawURLs {
		report := urlCheckReport{url: raw}
		plans, errs := buildPortCheckPlans([]string{raw})
		if len(errs) > 0 {
			for _, err := range errs {
				report.checks = append(report.checks, checkPlanResult{
					name:   "端口检测",
					ok:     false,
					detail: err.Error(),
				})
			}
			reports = append(reports, report)
			continue
		}

		if len(plans) == 0 {
			report.checks = append(report.checks, checkPlanResult{
				name:   "端口检测",
				ok:     true,
				detail: "未配置检测方案",
			})
			reports = append(reports, report)
			continue
		}

		for _, plan := range plans {
			if err := executePortCheck(plan); err != nil {
				report.checks = append(report.checks, checkPlanResult{
					name:   "端口检测",
					ok:     false,
					detail: err.Error(),
				})
				continue
			}
			report.checks = append(report.checks, checkPlanResult{
				name:   "端口检测",
				ok:     true,
				detail: "端口可用",
			})
		}
		reports = append(reports, report)
	}
	return reports
}
