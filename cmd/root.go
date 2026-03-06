/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
