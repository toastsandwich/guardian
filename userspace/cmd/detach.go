package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toastsandwich/guardian/internal/daemon"
)

var detachCmd = &cobra.Command{
	Use:   "detach",
	Short: "detach guardian from the interface",
	RunE: func(cmd *cobra.Command, args []string) error {
		code, err := daemon.DialDetach()
		if err != nil {
			return err
		}

		if code == daemon.CodeOK {
			fmt.Println("detached")
			return nil
		}
		fmt.Println("something went wrong check logs")
		return nil
	},
}
