package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func runScouts(cmd *cobra.Command, args []string) error {
	cfg, err := loadScoutConfig()
	if err != nil {
		printValidationError(cmd, err)
		return err
	}
	return runScoutsWithConfig(cmd, args, cfg)
}

func runScoutsWithConfig(cmd *cobra.Command, args []string, cfg scoutConfig) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	invalidURLs := validateConnectionURLs(args)
	if len(invalidURLs) > 0 {
		for _, err := range invalidURLs {
			printValidationError(cmd, err)
		}
		printValidationSummary(cmd, len(args), len(invalidURLs))
		return fmt.Errorf("one or more URLs are invalid")
	}

	systemDNS, _ := detectSystemDNS()
	systemDNSDisplay := "unknown"
	if len(systemDNS) > 0 {
		systemDNSDisplay = strings.Join(systemDNS, ", ")
	}
	fmt.Fprintf(cmd.OutOrStdout(), "[SYSTEM]\n🔍 当前DNS：%s\n", systemDNSDisplay)

	portReports := executePortChecks(args)
	dnsReports := executeDNSChecks(args, cfg.DNS)

	for i, report := range portReports {
		checks := report.checks
		if i < len(dnsReports) && dnsReports[i].url == report.url {
			checks = append(checks, dnsReports[i].checks...)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "\n[%s]\n", report.url)
		for _, check := range checks {
			mark := "✅"
			if !check.ok {
				mark = "❌"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s - %s\n", mark, check.name, check.detail)
		}
	}
	return nil
}

func printValidationError(cmd *cobra.Command, err error) {
	const redErrorPrefix = "\x1b[31m[ERROR]\x1b[0m"
	fmt.Fprintf(cmd.ErrOrStderr(), "%s %s\n", redErrorPrefix, err)
}

func printValidationSummary(cmd *cobra.Command, total, invalid int) {
	fmt.Fprintf(cmd.ErrOrStderr(), "Summary: total=%d, invalid=%d\n", total, invalid)
}
