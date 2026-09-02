package main

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/toastsandwich/guardian/internal/daemon"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "list stored ips and their status",
	RunE: func(cmd *cobra.Command, args []string) error {
		ips, code, err := daemon.NewClient().List(daemon.ListOptions{
			Limit: 10,
		})
		if err != nil {
			return err
		}
		if code != daemon.CodeOK {
			return fmt.Errorf("something went wrong check error logs")
		}

		builder := strings.Builder{}
		status := func(p bool) string {
			if p {
				return "Allowed"
			}
			return "Denied"
		}
		tw := tabwriter.NewWriter(&builder, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "\nIP\tSTATUS")
		fmt.Fprintln(tw, "--\t------")
		for _, ip := range ips.IPs {
			fmt.Fprintf(tw, "%s\t%s\n", ip.IP, status(ip.Allow))
		}
		tw.Flush()
		fmt.Print(builder.String())
		return nil
	},
}
