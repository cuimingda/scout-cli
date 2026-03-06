package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

func runScouts(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

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

	if len(args) > 1 {
		err := fmt.Errorf("scout一次只能处理一个输入，当前收到%d个", len(args))
		printValidationError(cmd, err)
		return err
	}

	raw := args[0]
	target, formatCheck := executeFormatCheck(raw)
	if !formatCheck.ok {
		fmt.Fprintf(cmd.OutOrStdout(), "[%s]\n", raw)
		writeCheckLine(cmd.OutOrStdout(), formatCheck)
		return fmt.Errorf("format check failed")
	}

	systemDNS, _ := detectSystemDNS()
	systemDNSDisplay := "unknown"
	if len(systemDNS) > 0 {
		systemDNSDisplay = strings.Join(systemDNS, ", ")
	}
	fmt.Fprintf(cmd.OutOrStdout(), "[SYSTEM]\n🔍 当前DNS：%s\n", systemDNSDisplay)

	fmt.Fprintf(cmd.OutOrStdout(), "\n[%s]\n", raw)
	writeCheckLine(cmd.OutOrStdout(), formatCheck)
	executePortChecksStreaming(target, func(check checkPlanResult) {
		writeCheckLine(cmd.OutOrStdout(), check)
	})
	executeDNSChecksStreamingWithResolvers(target, cfg.DNS, func(check checkPlanResult) {
		writeCheckLine(cmd.OutOrStdout(), check)
	})
	return nil
}

func writeCheckLine(out io.Writer, check checkPlanResult) {
	mark := "✅"
	if !check.ok {
		mark = "❌"
	}
	fmt.Fprintf(out, "%s %s - %s\n", mark, check.name, check.detail)
}

func printValidationError(cmd *cobra.Command, err error) {
	const redErrorPrefix = "\x1b[31m[ERROR]\x1b[0m"
	fmt.Fprintf(cmd.ErrOrStderr(), "%s %s\n", redErrorPrefix, err)
}

func printValidationSummary(cmd *cobra.Command, total, invalid int) {
	fmt.Fprintf(cmd.ErrOrStderr(), "Summary: total=%d, invalid=%d\n", total, invalid)
}
