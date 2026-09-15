package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "at-inspect",
	Short: "CLI tool for exploring the AT Protocol / Bluesky network",
}

func Execute() error {
	rootCmd.AddCommand(resolveCmd)
	// rootCmd.AddCommand(didCmd, repoCmd, recordCmd, firehoseCmd) — sırayla ekleyeceksin
	return rootCmd.Execute()
}
