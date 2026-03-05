package cmd

import (
	"bytes"
	"strings"
	"testing"

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
	cmd := &cobra.Command{Use: "scout [urls...]"}
	cmd.SetOut(&bytes.Buffer{})

	t.Run("prints each valid url", func(t *testing.T) {
		var out bytes.Buffer
		cmd := &cobra.Command{Use: "scout [urls...]"}
		cmd.SetOut(&out)

		err := runScouts(cmd, []string{
			"https://www.google.com/sitemap.xml",
			"udp://tracker.opentrackr.org:1337/announce",
		})
		if err != nil {
			t.Fatalf("runScouts() error = %v", err)
		}

		got := strings.TrimSpace(out.String())
		want := strings.Join([]string{
			"https://www.google.com/sitemap.xml",
			"udp://tracker.opentrackr.org:1337/announce",
		}, "\n")
		if got != want {
			t.Fatalf("runScouts output=%q want=%q", got, want)
		}
	})

	t.Run("returns error on invalid", func(t *testing.T) {
		var out bytes.Buffer
		cmd := &cobra.Command{Use: "scout [urls...]"}
		cmd.SetOut(&out)

		err := runScouts(cmd, []string{"google.com", "https://www.google.com"})
		if err == nil {
			t.Fatal("runScouts() expected error")
		}

		if strings.TrimSpace(out.String()) != "" {
			t.Fatalf("expected no output, got: %q", out.String())
		}
	})

	t.Run("shows help when no args", func(t *testing.T) {
		var out bytes.Buffer
		errBuf := bytes.Buffer{}
		cmd := &cobra.Command{Use: "scout [urls...]", Short: "test command"}
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
