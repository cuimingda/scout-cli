package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func runScouts(cmd *cobra.Command, args []string) error {
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

	for _, report := range executePortChecks(args) {
		fmt.Fprintf(cmd.OutOrStdout(), "\n[%s]\n", report.url)
		for _, check := range report.checks {
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
