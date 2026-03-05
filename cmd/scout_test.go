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
		defer func() { detectDial = origDial }()
		detectDial = func(string, string, time.Duration) error { return nil }

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
[http://bdremux.club/announce]
✅ 端口检测 - 端口可用
[udp://tracker.opentrackr.org:1337/announce]
✅ 端口检测 - 端口可用
[https://www.google.com/]
✅ 端口检测 - 端口可用`)
		if got != want {
			t.Fatalf("runScouts output=%q want=%q", got, want)
		}
	})

	t.Run("prints mixed check results", func(t *testing.T) {
		var out bytes.Buffer
		cmd := &cobra.Command{Use: "scout [urls...]"}
		cmd.SetOut(&out)
		origDial := detectDial
		defer func() { detectDial = origDial }()
		detectDial = func(string, string, time.Duration) error {
			return fmt.Errorf("simulated connect failed")
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
[http://bdremux.club/announce]
❌ 端口检测 - port check failed for "http://bdremux.club/announce" (bdremux.club:80 via tcp): simulated connect failed
[udp://tracker.opentrackr.org:1337/announce]
❌ 端口检测 - port check failed for "udp://tracker.opentrackr.org:1337/announce" (tracker.opentrackr.org:1337 via udp): simulated connect failed`)
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
