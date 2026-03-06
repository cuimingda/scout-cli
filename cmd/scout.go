package cmd

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cuimingda/scout-cli/internal/checker"
	"github.com/spf13/cobra"
)

var buildScoutManager = func(cfg scoutConfig) checker.Manager {
	formatChecker := checker.NewFormatChecker()
	portChecker := checker.NewPortChecker(checker.PortCheckerOptions{
		Timeout: 3 * time.Second,
	})
	dnsChecker := checker.NewDNSChecker(checker.DNSCheckerOptions{
		ExtraResolvers: cfg.DNS,
		Timeout:        3 * time.Second,
	})

	return checker.NewManager(
		formatChecker,
		dnsChecker,
		[]checker.Checker{dnsChecker},
		map[string][]checker.Checker{
			"http":  []checker.Checker{portChecker, dnsChecker},
			"https": []checker.Checker{portChecker, dnsChecker},
			"udp":   []checker.Checker{portChecker, dnsChecker},
		},
	)
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
	manager := buildScoutManager(cfg)
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

	systemDNS, _ := manager.SystemDNSes()
	systemDNSDisplay := "unknown"
	if len(systemDNS) > 0 {
		systemDNSDisplay = strings.Join(systemDNS, ", ")
	}
	fmt.Fprintf(cmd.OutOrStdout(), "[SYSTEM]\n🔍 当前DNS：%s\n", systemDNSDisplay)

	fmt.Fprintf(cmd.OutOrStdout(), "\n[%s]\n", raw)
	for _, result := range results {
		writeCheckLine(cmd.OutOrStdout(), result)
	}
	return nil
}

func writeCheckLine(out io.Writer, check checker.Result) {
	mark := "✅"
	if !check.OK {
		mark = "❌"
	}
	fmt.Fprintf(out, "%s %s - %s\n", mark, check.Name, check.Detail)
}

func printValidationError(cmd *cobra.Command, err error) {
	const redErrorPrefix = "\x1b[31m[ERROR]\x1b[0m"
	fmt.Fprintf(cmd.ErrOrStderr(), "%s %s\n", redErrorPrefix, err)
}

func printValidationSummary(cmd *cobra.Command, total, invalid int) {
	fmt.Fprintf(cmd.ErrOrStderr(), "Summary: total=%d, invalid=%d\n", total, invalid)
}
