package main

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use: "guardian",
}

func init() {
	rootCmd.AddCommand(attachCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(pingCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(stopCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
