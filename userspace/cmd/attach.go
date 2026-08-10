package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toastsandwich/guardian/internal/daemon"
)

var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "attach guardian to interface",
	RunE: func(cmd *cobra.Command, args []string) error {
		to, err := cmd.Flags().GetString("to")
		if err != nil {
			return err
		}
		if to == "" {
			cmd.Usage()
			return fmt.Errorf("--to flag required")
		}

		resp := daemon.AttachDial(daemon.Request{
			Command:   daemon.Attach,
			IfaceName: to,
		})

		if resp.Status == daemon.StatusAttached {
			fmt.Println("attached")
		} else {
			fmt.Println("something went wrong check logs")
		}

		return resp.Error
	},
}

func init() {
	attachCmd.Flags().StringP("to", "T", "", "used to provide name of interface to be attached to")
	attachCmd.MarkFlagRequired("to")
}
