package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/cuimingda/scout-cli/internal/checker"
	"github.com/spf13/cobra"
)

var buildScoutManager = func(cfg scoutConfig, dnsEnabled, portEnabled bool) checker.Manager {
	formatChecker := checker.NewFormatChecker()
	var portChecker *checker.PortChecker
	var dnsChecker *checker.DNSChecker
	if portEnabled {
		portChecker = checker.NewPortChecker(checker.PortCheckerOptions{
			Timeout: 3 * time.Second,
		})
	}
	if dnsEnabled {
		dnsChecker = checker.NewDNSChecker(checker.DNSCheckerOptions{
			ExtraResolvers: cfg.DNS,
			Timeout:        3 * time.Second,
		})
	}

	return checker.NewManager(
		formatChecker,
		defaultProtocolCheckers(dnsChecker),
		map[string][]checker.Checker{
			"http":  protocolCheckersFor(portChecker, dnsChecker),
			"https": protocolCheckersFor(portChecker, dnsChecker),
			"udp":   protocolCheckersFor(portChecker, dnsChecker),
		},
	)
}

var buildSystemManager = func(systemEnabled bool) checker.SystemManager {
	if !systemEnabled {
		return checker.NewSystemManager(nil)
	}
	return checker.NewSystemManager([]checker.SystemChecker{
		checker.NewSystemDNSChecker(checker.SystemDNSCheckerOptions{}),
	})
}

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
	dnsEnabled, portEnabled, systemEnabled := enabledOptionalChecks()
	systemManager := buildSystemManager(systemEnabled)
	_, systemResults := systemManager.Run()
	if len(systemResults) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "[SYSTEM]\n")
		for _, result := range systemResults {
			writeSystemLine(cmd.OutOrStdout(), result)
		}
		fmt.Fprint(cmd.OutOrStdout(), "\n")
	}

	manager := buildScoutManager(cfg, dnsEnabled, portEnabled)
	_, results := manager.Run(raw)
	if len(results) == 0 {
		return fmt.Errorf("no checks executed")
	}
	if !results[0].OK {
		fmt.Fprintf(cmd.OutOrStdout(), "[%s]\n", raw)
		for _, result := range results {
			writeCheckLine(cmd.OutOrStdout(), result)
		}
		return fmt.Errorf("format check failed")
	}

	fmt.Fprintf(cmd.OutOrStdout(), "[%s]\n", raw)
	for _, result := range results {
		writeCheckLine(cmd.OutOrStdout(), result)
	}
	return nil
}

func enabledOptionalChecks() (bool, bool, bool) {
	if scoutAllChecksEnabled {
		return true, true, true
	}
	return scoutDNSCheckEnabled, scoutPortCheckEnabled, scoutSystemCheckEnabled
}

func defaultProtocolCheckers(dnsChecker *checker.DNSChecker) []checker.Checker {
	if dnsChecker == nil {
		return nil
	}
	return []checker.Checker{dnsChecker}
}

func protocolCheckersFor(portChecker *checker.PortChecker, dnsChecker *checker.DNSChecker) []checker.Checker {
	checkers := make([]checker.Checker, 0, 2)
	if portChecker != nil {
		checkers = append(checkers, portChecker)
	}
	if dnsChecker != nil {
		checkers = append(checkers, dnsChecker)
	}
	return checkers
}

func writeCheckLine(out io.Writer, check checker.Result) {
	mark := "✅"
	if !check.OK {
		mark = "❌"
	}
	fmt.Fprintf(out, "%s %s - %s\n", mark, check.Name, check.Detail)
}

func writeSystemLine(out io.Writer, check checker.Result) {
	if check.OK {
		fmt.Fprintf(out, "🔍 %s：%s\n", check.Name, check.Detail)
		return
	}
	fmt.Fprintf(out, "❌ %s - %s\n", check.Name, check.Detail)
}

func printValidationError(cmd *cobra.Command, err error) {
	const redErrorPrefix = "\x1b[31m[ERROR]\x1b[0m"
	fmt.Fprintf(cmd.ErrOrStderr(), "%s %s\n", redErrorPrefix, err)
}

func printValidationSummary(cmd *cobra.Command, total, invalid int) {
	fmt.Fprintf(cmd.ErrOrStderr(), "Summary: total=%d, invalid=%d\n", total, invalid)
}
