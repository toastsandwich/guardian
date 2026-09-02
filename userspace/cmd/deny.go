package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toastsandwich/guardian/internal/daemon"
)

var denyCmd = &cobra.Command{
	Use:   "deny",
	Short: "deny an ip address",
	Long: `denies an ip address from sending requests to the machine guardian is hosted on.
Currently only ipv4 is supported.
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ip := args[0]
		code, err := daemon.NewClient().Deny(daemon.DenyOptions{
			IP: ip,
		})
		if err != nil {
			return err
		}
		if code != daemon.CodeOK {
			return fmt.Errorf("something went wrong check guardian logs")
		}
		return nil
	},
}
