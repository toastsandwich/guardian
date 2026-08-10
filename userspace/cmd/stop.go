package main

import (
	"github.com/spf13/cobra"
	"github.com/toastsandwich/guardian/internal/daemon"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stops the guardian daemon (recommended to stop the service directly)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon.StopDial(daemon.Request{Command: daemon.Stop})
	},
}
