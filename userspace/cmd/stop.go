package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toastsandwich/guardian/internal/daemon"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "detach from the interface and stop the guardian daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		code, err := daemon.NewClient().Stop()
		if err != nil {
			return err
		}
		if code != daemon.CodeOK {
			fmt.Println("something went wrong check guardian logs")
			return nil
		}
		fmt.Println("detached and stopped")
		return nil
	},
}
