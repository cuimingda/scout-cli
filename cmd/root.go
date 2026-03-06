/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var scoutDNSCheckEnabled bool
var scoutPortCheckEnabled bool
var scoutSystemCheckEnabled bool
var scoutAllChecksEnabled bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "scout <url>",
	Short:   "Scout a protocol-based connection",
	Long:    "Scout validates one complete protocol URL and runs format, port, and DNS checks for that input.",
	Example: `scout https://www.google.com/sitemap.xml
scout udp://tracker.opentrackr.org:1337/announce`,
	RunE:        runScouts,
	SilenceUsage: true,
	SilenceErrors: true,
}

func init() {
	rootCmd.Flags().BoolVar(&scoutDNSCheckEnabled, "dns", false, "Enable DNS checks")
	rootCmd.Flags().BoolVar(&scoutPortCheckEnabled, "port", false, "Enable port checks")
	rootCmd.Flags().BoolVar(&scoutSystemCheckEnabled, "system", false, "Show system information")
	rootCmd.Flags().BoolVar(&scoutAllChecksEnabled, "all", false, "Enable all optional checks")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
