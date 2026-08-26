package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toastsandwich/guardian/internal/daemon"
)

var d *daemon.Server

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "test command to see unix server",
	RunE: func(cmd *cobra.Command, args []string) error {
		if d == nil {
			fmt.Println("here")
			d = daemon.NewServer()
		}
		return d.Start()
	},
}
