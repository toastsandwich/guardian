package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toastsandwich/guardian/internal/daemon"
)

var modeCmd = &cobra.Command{
	Use:   "mode",
	Short: "get or set guardian mode",
	Long: `gets or sets the guardian filtering mode.

monk   allow by default, drop ips added with deny
sentry deny by default, pass ips added with allow`,
}

var modeGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get current guardian mode",
	Long:  `prints the current guardian filtering mode.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mode, code, err := daemon.NewClient().GetMode()
		if err != nil {
			return err
		}
		if code != daemon.CodeOK {
			return fmt.Errorf("something went wrong check guardian logs")
		}
		fmt.Println(mode)
		return nil
	},
}

var modeSetCmd = &cobra.Command{
	Use:   "set [mode]",
	Short: "set guardian mode",
	Long: `sets the guardian filtering mode.

monk   allow by default, drop ips added with deny
sentry deny by default, pass ips added with allow`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		code, err := daemon.NewClient().SetMode(daemon.SetModeOptions{
			Mode: args[0],
		})
		if err != nil {
			return err
		}
		if code != daemon.CodeOK {
			return fmt.Errorf("something went wrong check guardian logs")
		}
		fmt.Println("mode set")
		return nil
	},
}

func init() {
	modeCmd.AddCommand(modeGetCmd)
	modeCmd.AddCommand(modeSetCmd)
}
