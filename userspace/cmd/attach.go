package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toastsandwich/guardian/internal/daemon"
)

var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "attach guardian to interface",
	Long: `attaches guardian to a network interface and starts filtering incoming ipv4 traffic.

--to is the interface name (for example eth0).
--mode selects how unknown ips are treated:
  monk   allow by default, drop ips added with deny
  sentry deny by default, pass ips added with allow`,
	RunE: func(cmd *cobra.Command, args []string) error {
		to, err := cmd.Flags().GetString("to")
		if err != nil {
			return err
		}

		mode, err := cmd.Flags().GetString("mode")
		if err != nil {
			return err
		}

		code, err := daemon.NewClient().Attach(daemon.AttachOptions{
			IfaceName: to,
			Mode:      mode,
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
	attachCmd.Flags().StringP("mode", "M", "", "used to provide mode of guardian")
	attachCmd.MarkFlagRequired("to")
	attachCmd.MarkFlagRequired("mode")
}
