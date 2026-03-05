package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func runScouts(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	for _, raw := range args {
		if err := validateConnectionURL(raw); err != nil {
			err = fmt.Errorf("invalid URL %q: %w", raw, err)
			printValidationError(cmd, err)
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), raw)
	}
	return nil
}

func printValidationError(cmd *cobra.Command, err error) {
	const redErrorPrefix = "\x1b[31m[ERROR]\x1b[0m"
	fmt.Fprintf(cmd.ErrOrStderr(), "%s %s\n", redErrorPrefix, err)
}
