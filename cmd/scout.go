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
			return fmt.Errorf("invalid URL %q: %w", raw, err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), raw)
	}
	return nil
}
