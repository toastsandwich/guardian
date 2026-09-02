package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/toastsandwich/guardian/internal/daemon"
)

var d *daemon.Server

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "test command to see unix server",
	RunE: func(cmd *cobra.Command, args []string) error {
		dev, err := cmd.Flags().GetBool("dev")
		if err != nil {
			return err
		}
		if dev {
			slog.SetLogLoggerLevel(slog.LevelDebug)
		}

		if d == nil {
			fmt.Println("here")
			d = daemon.NewServer()
		}
		return d.Start()
	},
}

func init() {
	startCmd.Flags().Bool("dev", false, "enable debug logs")
	startCmd.Flags().MarkHidden("dev")
}
