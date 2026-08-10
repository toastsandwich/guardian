package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toastsandwich/guardian/internal/daemon"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "shows stored ips and its status",
	Run: func(cmd *cobra.Command, args []string) {
		resp := daemon.ShowDial(daemon.Request{
			Command: daemon.Show,
		})
		if resp.Ips != nil {
			for _, ip := range resp.Ips {
				fmt.Printf("%s:%10v\n", ip.IP, ip.Allowed)
			}
		}
	},
}
