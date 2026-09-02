package main

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use: "guardian",
}

func init() {
	rootCmd.AddCommand(attachCmd)
	rootCmd.AddCommand(detachCmd)
	rootCmd.AddCommand(startCmd)
	// rootCmd.AddCommand(pingCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(allowCmd)
	rootCmd.AddCommand(denyCmd)
	rootCmd.AddCommand(modeCmd)
	rootCmd.AddCommand(stopCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
