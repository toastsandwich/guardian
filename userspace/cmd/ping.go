package main

// var pingCmd = &cobra.Command{
// 	Use:   "ping",
// 	Short: "test command to check with server",
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		resp := daemon.PingDial(daemon.Request{
// 			Command: daemon.Ping,
// 		})
// 		if resp.Error != nil {
// 			return resp.Error
// 		}

// 		if resp.Status == daemon.StatusOK {
// 			fmt.Println("OK")
// 		} else {
// 			fmt.Println("NOT OK")
// 		}

// 		return nil
// 	},
// }
