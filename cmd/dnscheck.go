package cmd

import (
	"fmt"
	"strings"
)

func executeDNSChecks(rawURLs []string) []urlCheckReport {
	reports := make([]urlCheckReport, 0, len(rawURLs))
	for _, raw := range rawURLs {
		report := urlCheckReport{url: raw}
		plans, errs := buildDNSCheckPlans([]string{raw})
		if len(errs) > 0 {
			for _, err := range errs {
				report.checks = append(report.checks, checkPlanResult{
					name:   "DNS解析",
					ok:     false,
					detail: err.Error(),
				})
			}
			reports = append(reports, report)
			continue
		}

		for _, plan := range plans {
			ips, err := executeDNSCheck(plan)
			if err != nil {
				report.checks = append(report.checks, checkPlanResult{
					name:   "DNS解析",
					ok:     false,
					detail: err.Error(),
				})
				continue
			}
			report.checks = append(report.checks, checkPlanResult{
				name:   "DNS解析",
				ok:     true,
				detail: fmt.Sprintf("%s在%s解析到%s", plan.host, plan.resolverLabel, strings.Join(ips, ",")),
			})
		}

		reports = append(reports, report)
	}
	return reports
}
