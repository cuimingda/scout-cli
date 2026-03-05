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
	Use:     "scout [urls...]",
	Short:   "Scout one or more protocol-based connections",
	Long:    "Scout validates one or more complete protocol URLs and prints each valid input.",
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
