package main

import (
	"github.com/spf13/cobra"
	"github.com/toastsandwich/guardian/internal/daemon"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "test command to see unix server",
	RunE: func(cmd *cobra.Command, args []string) error {
		d := daemon.NewServer()
		return d.ListenAndServe()
	},
}
