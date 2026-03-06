package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cuimingda/scout-cli/internal/checker"
	"github.com/spf13/cobra"
)

func buildTestManager(cfg scoutConfig, dialErr error, dnsErr error, dnsEnabled bool, portEnabled bool) checker.Manager {
	formatChecker := checker.NewFormatChecker()
	var portChecker *checker.PortChecker
	var dnsChecker *checker.DNSChecker
	if portEnabled {
		portChecker = checker.NewPortChecker(checker.PortCheckerOptions{
			Timeout: time.Second,
			Dial: func(string, string, time.Duration) error {
				return dialErr
			},
		})
	}
	if dnsEnabled {
		dnsChecker = checker.NewDNSChecker(checker.DNSCheckerOptions{
			ExtraResolvers: cfg.DNS,
			Timeout:        time.Second,
			SystemDNS: func() ([]string, error) {
				return []string{"8.8.8.8", "223.5.5.5"}, nil
			},
			Lookup: func(host, resolver string, _ time.Duration) ([]string, error) {
				if dnsErr != nil {
					return nil, dnsErr
				}
				return []string{"1.2.3.4"}, nil
			},
		})
	}

	return checker.NewManager(
		formatChecker,
		dnsChecker,
		defaultProtocolCheckers(dnsChecker),
		map[string][]checker.Checker{
			"http":  protocolCheckersFor(portChecker, dnsChecker),
			"https": protocolCheckersFor(portChecker, dnsChecker),
			"udp":   protocolCheckersFor(portChecker, dnsChecker),
		},
	)
}

func setScoutFlags(dns bool, port bool, all bool) func() {
	prevDNS := scoutDNSCheckEnabled
	prevPort := scoutPortCheckEnabled
	prevAll := scoutAllChecksEnabled
	scoutDNSCheckEnabled = dns
	scoutPortCheckEnabled = port
	scoutAllChecksEnabled = all
	return func() {
		scoutDNSCheckEnabled = prevDNS
		scoutPortCheckEnabled = prevPort
		scoutAllChecksEnabled = prevAll
	}
}

func Test_runScouts(t *testing.T) {
	t.Run("prints only format result by default", func(t *testing.T) {
		var out bytes.Buffer
		cmd := &cobra.Command{Use: "scout [urls...]"}
		cmd.SetOut(&out)
		restoreFlags := setScoutFlags(false, false, false)
		defer restoreFlags()
		origBuild := buildScoutManager
		defer func() { buildScoutManager = origBuild }()
		buildScoutManager = func(cfg scoutConfig, dnsEnabled, portEnabled bool) checker.Manager {
			return buildTestManager(cfg, nil, nil, dnsEnabled, portEnabled)
		}

		err := runScoutsWithConfig(cmd, []string{
			"http://bdremux.club/announce",
		}, scoutConfig{DNS: []string{"223.5.5.5", "8.8.8.8"}})
		if err != nil {
			t.Fatalf("runScouts() error = %v", err)
		}

		got := strings.TrimSpace(out.String())
		want := strings.TrimSpace(`
[http://bdremux.club/announce]
✅ 文件格式检查 - 输入格式合法`)
		if got != want {
			t.Fatalf("runScouts output=%q want=%q", got, want)
		}
	})

	t.Run("prints dns results when dns flag enabled", func(t *testing.T) {
		var out bytes.Buffer
		cmd := &cobra.Command{Use: "scout [urls...]"}
		cmd.SetOut(&out)
		restoreFlags := setScoutFlags(true, false, false)
		defer restoreFlags()
		origBuild := buildScoutManager
		defer func() { buildScoutManager = origBuild }()
		buildScoutManager = func(cfg scoutConfig, dnsEnabled, portEnabled bool) checker.Manager {
			return buildTestManager(cfg, nil, nil, dnsEnabled, portEnabled)
		}

		err := runScoutsWithConfig(cmd, []string{
			"http://bdremux.club/announce",
		}, scoutConfig{DNS: []string{"223.5.5.5", "8.8.8.8"}})
		if err != nil {
			t.Fatalf("runScouts() error = %v", err)
		}

		got := strings.TrimSpace(out.String())
		want := strings.TrimSpace(`
[SYSTEM]
🔍 当前DNS：8.8.8.8, 223.5.5.5

[http://bdremux.club/announce]
✅ 文件格式检查 - 输入格式合法
✅ DNS解析 - bdremux.club在当前DNS解析到1.2.3.4
✅ DNS解析 - bdremux.club在223.5.5.5解析到1.2.3.4
✅ DNS解析 - bdremux.club在8.8.8.8解析到1.2.3.4`)
		if got != want {
			t.Fatalf("runScouts output=%q want=%q", got, want)
		}
	})

	t.Run("prints port results when port flag enabled", func(t *testing.T) {
		var out bytes.Buffer
		cmd := &cobra.Command{Use: "scout [urls...]"}
		cmd.SetOut(&out)
		restoreFlags := setScoutFlags(false, true, false)
		defer restoreFlags()
		origBuild := buildScoutManager
		defer func() { buildScoutManager = origBuild }()
		buildScoutManager = func(cfg scoutConfig, dnsEnabled, portEnabled bool) checker.Manager {
			return buildTestManager(cfg, fmt.Errorf("simulated connect failed"), nil, dnsEnabled, portEnabled)
		}

		err := runScoutsWithConfig(cmd, []string{
			"http://bdremux.club/announce",
		}, scoutConfig{DNS: []string{"223.5.5.5", "8.8.8.8"}})
		if err != nil {
			t.Fatalf("runScouts() error = %v", err)
		}

		got := strings.TrimSpace(out.String())
		want := strings.TrimSpace(`
[http://bdremux.club/announce]
✅ 文件格式检查 - 输入格式合法
❌ 端口检测 - bdremux.club的80端口未开放（simulated connect failed）`)
		if got != want {
			t.Fatalf("runScouts output=%q want=%q", got, want)
		}
	})

	t.Run("prints all results when all flag enabled", func(t *testing.T) {
		var out bytes.Buffer
		cmd := &cobra.Command{Use: "scout [urls...]"}
		cmd.SetOut(&out)
		restoreFlags := setScoutFlags(false, false, true)
		defer restoreFlags()
		origBuild := buildScoutManager
		defer func() { buildScoutManager = origBuild }()
		buildScoutManager = func(cfg scoutConfig, dnsEnabled, portEnabled bool) checker.Manager {
			return buildTestManager(cfg, nil, nil, dnsEnabled, portEnabled)
		}

		err := runScoutsWithConfig(cmd, []string{
			"http://bdremux.club/announce",
		}, scoutConfig{DNS: []string{"223.5.5.5", "8.8.8.8"}})
		if err != nil {
			t.Fatalf("runScouts() error = %v", err)
		}

		got := strings.TrimSpace(out.String())
		want := strings.TrimSpace(`
[SYSTEM]
🔍 当前DNS：8.8.8.8, 223.5.5.5

[http://bdremux.club/announce]
✅ 文件格式检查 - 输入格式合法
✅ 端口检测 - bdremux.club的80端口开放
✅ DNS解析 - bdremux.club在当前DNS解析到1.2.3.4
✅ DNS解析 - bdremux.club在223.5.5.5解析到1.2.3.4
✅ DNS解析 - bdremux.club在8.8.8.8解析到1.2.3.4`)
		if got != want {
			t.Fatalf("runScouts output=%q want=%q", got, want)
		}
	})

	t.Run("supports enabling dns and port together without all", func(t *testing.T) {
		var out bytes.Buffer
		cmd := &cobra.Command{Use: "scout [urls...]"}
		cmd.SetOut(&out)
		restoreFlags := setScoutFlags(true, true, false)
		defer restoreFlags()
		origBuild := buildScoutManager
		defer func() { buildScoutManager = origBuild }()
		buildScoutManager = func(cfg scoutConfig, dnsEnabled, portEnabled bool) checker.Manager {
			return buildTestManager(cfg, nil, nil, dnsEnabled, portEnabled)
		}

		err := runScoutsWithConfig(cmd, []string{
			"https://www.google.com/",
		}, scoutConfig{DNS: []string{"223.5.5.5", "8.8.8.8"}})
		if err != nil {
			t.Fatalf("runScouts() error = %v", err)
		}

		got := strings.TrimSpace(out.String())
		want := strings.TrimSpace(`
[SYSTEM]
🔍 当前DNS：8.8.8.8, 223.5.5.5

[https://www.google.com/]
✅ 文件格式检查 - 输入格式合法
✅ 端口检测 - www.google.com的443端口开放
✅ DNS解析 - www.google.com在当前DNS解析到1.2.3.4
✅ DNS解析 - www.google.com在223.5.5.5解析到1.2.3.4
✅ DNS解析 - www.google.com在8.8.8.8解析到1.2.3.4`)
		if got != want {
			t.Fatalf("runScouts output=%q want=%q", got, want)
		}
	})

	t.Run("returns error when format check fails", func(t *testing.T) {
		var out bytes.Buffer
		cmd := &cobra.Command{Use: "scout [urls...]"}
		cmd.SetOut(&out)
		restoreFlags := setScoutFlags(false, false, true)
		defer restoreFlags()
		origBuild := buildScoutManager
		defer func() { buildScoutManager = origBuild }()
		buildScoutManager = func(cfg scoutConfig, dnsEnabled, portEnabled bool) checker.Manager {
			return buildTestManager(cfg, nil, nil, dnsEnabled, portEnabled)
		}

		err := runScoutsWithConfig(cmd, []string{"google.com"}, scoutConfig{DNS: []string{"223.5.5.5", "8.8.8.8"}})
		if err == nil {
			t.Fatal("runScouts() expected error")
		}

		got := strings.TrimSpace(out.String())
		want := strings.TrimSpace(`
[google.com]
❌ 文件格式检查 - invalid URL "google.com": missing protocol`)
		if got != want {
			t.Fatalf("runScouts output=%q want=%q", got, want)
		}
	})

	t.Run("rejects multiple inputs", func(t *testing.T) {
		var out, errBuf bytes.Buffer
		cmd := &cobra.Command{Use: "scout [urls...]"}
		cmd.SetOut(&out)
		cmd.SetErr(&errBuf)
		restoreFlags := setScoutFlags(false, false, false)
		defer restoreFlags()

		err := runScoutsWithConfig(cmd, []string{"https://www.google.com", "https://www.github.com"}, scoutConfig{DNS: []string{"223.5.5.5", "8.8.8.8"}})
		if err == nil {
			t.Fatal("runScouts() expected error")
		}

		wantErr := fmt.Sprintf("\x1b[31m[ERROR]\x1b[0m scout一次只能处理一个输入，当前收到%d个\n", 2)
		if errBuf.String() != wantErr {
			t.Fatalf("runScouts() error output = %q, want = %q", errBuf.String(), wantErr)
		}
		if strings.TrimSpace(out.String()) != "" {
			t.Fatalf("expected no output, got: %q", out.String())
		}
	})

	t.Run("runs dns only for ftp when dns flag enabled", func(t *testing.T) {
		var out bytes.Buffer
		cmd := &cobra.Command{Use: "scout [urls...]"}
		cmd.SetOut(&out)
		restoreFlags := setScoutFlags(true, false, false)
		defer restoreFlags()
		origBuild := buildScoutManager
		defer func() { buildScoutManager = origBuild }()
		buildScoutManager = func(cfg scoutConfig, dnsEnabled, portEnabled bool) checker.Manager {
			return buildTestManager(cfg, nil, nil, dnsEnabled, portEnabled)
		}

		err := runScoutsWithConfig(cmd, []string{
			"ftp://ftp.example.com/resource",
		}, scoutConfig{DNS: []string{"223.5.5.5", "8.8.8.8"}})
		if err != nil {
			t.Fatalf("runScouts() error = %v", err)
		}

		got := strings.TrimSpace(out.String())
		want := strings.TrimSpace(`
[SYSTEM]
🔍 当前DNS：8.8.8.8, 223.5.5.5

[ftp://ftp.example.com/resource]
✅ 文件格式检查 - 输入格式合法
✅ DNS解析 - ftp.example.com在当前DNS解析到1.2.3.4
✅ DNS解析 - ftp.example.com在223.5.5.5解析到1.2.3.4
✅ DNS解析 - ftp.example.com在8.8.8.8解析到1.2.3.4`)
		if got != want {
			t.Fatalf("runScouts output=%q want=%q", got, want)
		}
	})

	t.Run("shows help when no args", func(t *testing.T) {
		var out bytes.Buffer
		errBuf := bytes.Buffer{}
		cmd := &cobra.Command{Use: "scout [urls]", Short: "test command"}
		cmd.SetOut(&out)
		cmd.SetErr(&errBuf)

		err := runScouts(cmd, []string{})
		if err != nil {
			t.Fatalf("runScouts() error = %v", err)
		}

		if !strings.Contains(out.String(), "test command") {
			t.Fatalf("help output not contain short description, got: %q", out.String())
		}
	})
}
