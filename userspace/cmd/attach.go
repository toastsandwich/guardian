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

		code, err := daemon.DialAttach(daemon.AttachOptions{
			IfaceName: to,
		})
		if err != nil {
			return err
		}

		if code == daemon.CodeOK {
			fmt.Println("attached")
			return nil
		}
		fmt.Println("something went wrong check logs")
		return nil
	},
}

func init() {
	attachCmd.Flags().StringP("to", "T", "", "used to provide name of interface to be attached to")
	attachCmd.MarkFlagRequired("to")
}
