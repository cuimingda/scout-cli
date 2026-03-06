package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func mustBuildScoutTarget(t *testing.T, raw string) scoutTarget {
	t.Helper()

	target, err := buildScoutTarget(raw)
	if err != nil {
		t.Fatalf("buildScoutTarget(%q) error = %v", raw, err)
	}
	return target
}

func Test_validateConnectionURL(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{name: "valid_http_path", input: "https://www.google.com/sitemap.xml", wantError: false},
		{name: "valid_udp", input: "udp://tracker.opentrackr.org:1337/announce", wantError: false},
		{name: "valid_ftp", input: "ftp://ftp.example.com/resource", wantError: false},
		{name: "missing_scheme", input: "google.com", wantError: true},
		{name: "missing_host", input: "https://", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConnectionURL(tt.input)
			if (err != nil) != tt.wantError {
				t.Fatalf("validateConnectionURL(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
			}
		})
	}
}

func Test_executeFormatCheck(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		target, check := executeFormatCheck("https://www.google.com/sitemap.xml")
		if !check.ok {
			t.Fatalf("expected success, got %+v", check)
		}
		if check.name != "文件格式检查" || check.detail != "输入格式合法" {
			t.Fatalf("unexpected check: %+v", check)
		}
		if target.raw != "https://www.google.com/sitemap.xml" || target.parsed == nil {
			t.Fatalf("unexpected target: %+v", target)
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		_, check := executeFormatCheck("google.com")
		if check.ok {
			t.Fatalf("expected failure, got %+v", check)
		}
		if check.name != "文件格式检查" {
			t.Fatalf("unexpected check name: %+v", check)
		}
		if !strings.Contains(check.detail, "invalid URL \"google.com\": missing protocol") {
			t.Fatalf("unexpected detail: %s", check.detail)
		}
	})
}

func Test_runScouts(t *testing.T) {
	t.Run("prints format, port and DNS results for single input", func(t *testing.T) {
		var out bytes.Buffer
		cmd := &cobra.Command{Use: "scout [urls...]"}
		cmd.SetOut(&out)
		origDial := detectDial
		origDNSLookup := detectDNSLookup
		origSystemDNS := detectSystemDNS
		defer func() { detectDial = origDial }()
		defer func() { detectDNSLookup = origDNSLookup }()
		defer func() { detectSystemDNS = origSystemDNS }()
		detectDial = func(string, string, time.Duration) error { return nil }
		detectSystemDNS = func() ([]string, error) {
			return []string{"8.8.8.8", "223.5.5.5"}, nil
		}
		detectDNSLookup = func(host, resolver string, _ time.Duration) ([]string, error) {
			return []string{"1.2.3.4"}, nil
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

	t.Run("prints mixed check results", func(t *testing.T) {
		var out bytes.Buffer
		cmd := &cobra.Command{Use: "scout [urls...]"}
		cmd.SetOut(&out)
		origDial := detectDial
		origDNSLookup := detectDNSLookup
		origSystemDNS := detectSystemDNS
		defer func() { detectDial = origDial }()
		defer func() { detectDNSLookup = origDNSLookup }()
		defer func() { detectSystemDNS = origSystemDNS }()
		detectDial = func(string, string, time.Duration) error {
			return fmt.Errorf("simulated connect failed")
		}
		detectSystemDNS = func() ([]string, error) {
			return []string{"8.8.8.8", "223.5.5.5"}, nil
		}
		detectDNSLookup = func(host, resolver string, _ time.Duration) ([]string, error) {
			return []string{"1.2.3.4"}, nil
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
❌ 端口检测 - bdremux.club的80端口未开放（simulated connect failed）
✅ DNS解析 - bdremux.club在当前DNS解析到1.2.3.4
✅ DNS解析 - bdremux.club在223.5.5.5解析到1.2.3.4
✅ DNS解析 - bdremux.club在8.8.8.8解析到1.2.3.4`)
		if got != want {
			t.Fatalf("runScouts output=%q want=%q", got, want)
		}
	})

	t.Run("returns error when format check fails", func(t *testing.T) {
		var out bytes.Buffer
		cmd := &cobra.Command{Use: "scout [urls...]"}
		cmd.SetOut(&out)

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
