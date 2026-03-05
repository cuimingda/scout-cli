package cmd

func executeDNSChecks(rawURLs []string, extraDNS []string) []urlCheckReport {
	return executeDNSChecksWithResolvers(rawURLs, extraDNS)
}
