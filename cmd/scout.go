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
		return fmt.Errorf("one or more URLs are invalid")
	}

	for _, raw := range args {
		fmt.Fprintln(cmd.OutOrStdout(), raw)
	}
	return nil
}

func printValidationError(cmd *cobra.Command, err error) {
	const redErrorPrefix = "\x1b[31m[ERROR]\x1b[0m"
	fmt.Fprintf(cmd.ErrOrStderr(), "%s %s\n", redErrorPrefix, err)
}
