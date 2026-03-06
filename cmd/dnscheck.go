package cmd

func executeDNSChecks(target scoutTarget, extraDNS []string) []checkPlanResult {
	return executeDNSChecksWithResolvers(target, extraDNS)
}
