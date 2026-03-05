package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

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

func Test_runScouts(t *testing.T) {
	t.Run("prints grouped check results", func(t *testing.T) {
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

		err := runScouts(cmd, []string{
			"http://bdremux.club/announce",
			"udp://tracker.opentrackr.org:1337/announce",
			"https://www.google.com/",
		})
		if err != nil {
			t.Fatalf("runScouts() error = %v", err)
		}

		got := strings.TrimSpace(out.String())
		want := strings.TrimSpace(`
[SYSTEM]
🔍 当前DNS：8.8.8.8, 223.5.5.5

[http://bdremux.club/announce]
✅ 端口检测 - bdremux.club的80端口开放
✅ DNS解析 - bdremux.club在当前DNS解析到1.2.3.4
✅ DNS解析 - bdremux.club在8.8.8.8解析到1.2.3.4
✅ DNS解析 - bdremux.club在223.5.5.5解析到1.2.3.4

[udp://tracker.opentrackr.org:1337/announce]
✅ 端口检测 - tracker.opentrackr.org的1337端口开放
✅ DNS解析 - tracker.opentrackr.org在当前DNS解析到1.2.3.4
✅ DNS解析 - tracker.opentrackr.org在8.8.8.8解析到1.2.3.4
✅ DNS解析 - tracker.opentrackr.org在223.5.5.5解析到1.2.3.4

[https://www.google.com/]
✅ 端口检测 - www.google.com的443端口开放
✅ DNS解析 - www.google.com在当前DNS解析到1.2.3.4
✅ DNS解析 - www.google.com在8.8.8.8解析到1.2.3.4
✅ DNS解析 - www.google.com在223.5.5.5解析到1.2.3.4`)
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

		err := runScouts(cmd, []string{
			"http://bdremux.club/announce",
			"udp://tracker.opentrackr.org:1337/announce",
		})
		if err != nil {
			t.Fatalf("runScouts() error = %v", err)
		}

		got := strings.TrimSpace(out.String())
		want := strings.TrimSpace(`
[SYSTEM]
🔍 当前DNS：8.8.8.8, 223.5.5.5

[http://bdremux.club/announce]
❌ 端口检测 - bdremux.club的80端口未开放（simulated connect failed）
✅ DNS解析 - bdremux.club在当前DNS解析到1.2.3.4
✅ DNS解析 - bdremux.club在8.8.8.8解析到1.2.3.4
✅ DNS解析 - bdremux.club在223.5.5.5解析到1.2.3.4

[udp://tracker.opentrackr.org:1337/announce]
❌ 端口检测 - tracker.opentrackr.org的1337端口未开放（simulated connect failed）
✅ DNS解析 - tracker.opentrackr.org在当前DNS解析到1.2.3.4
✅ DNS解析 - tracker.opentrackr.org在8.8.8.8解析到1.2.3.4
✅ DNS解析 - tracker.opentrackr.org在223.5.5.5解析到1.2.3.4`)
		if got != want {
			t.Fatalf("runScouts output=%q want=%q", got, want)
		}
	})

	t.Run("returns all errors when args invalid", func(t *testing.T) {
		var out, errBuf bytes.Buffer
		cmd := &cobra.Command{Use: "scout [urls...]"}
		cmd.SetOut(&out)
		cmd.SetErr(&errBuf)

		err := runScouts(cmd, []string{"google.com", "https://", "https://www.google.com"})
		if err == nil {
			t.Fatal("runScouts() expected error")
		}

		wantErr := strings.Join([]string{
			fmt.Sprintf("\x1b[31m[ERROR]\x1b[0m invalid URL %q: missing protocol", "google.com"),
			fmt.Sprintf("\x1b[31m[ERROR]\x1b[0m invalid URL %q: missing host", "https://"),
			"Summary: total=3, invalid=2",
		}, "\n") + "\n"
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
