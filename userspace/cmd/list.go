package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/toastsandwich/guardian/internal/daemon"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "list stored ips and their status",
	RunE: func(cmd *cobra.Command, args []string) error {
		builder := strings.Builder{}
		code, err := daemon.DialList(&builder, daemon.ListOptions{
			Limit: 10,
		})
		if err != nil {
			return err
		}
		if code != daemon.CodeOK {
			return fmt.Errorf("something went wrong check error logs")
		}
		fmt.Print(builder.String())
		return nil
	},
}
