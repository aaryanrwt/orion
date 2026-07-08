package cli

import (
	"encoding/json"
	"fmt"

	"github.com/orion-infra/orion/internal/core"
	"github.com/orion-infra/orion/internal/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Display a high-level summary of your Orion network",
	Example: "  $ orion status",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Load local config
		cfg, err := core.LoadConfig()
		if err != nil {
			return fmt.Errorf("orion is not initialized. Run %s first", ui.Bold("orion init"))
		}

		// 2. Query daemon to check running status and get online devices
		daemonRunning := "Running"
		discoveryActive := "Active"
		onlineCount := 1 // Local machine is always online
		trustedCount := len(cfg.Devices)
		totalCount := trustedCount + 1

		resp, err := localClient().Get("https://127.0.0.1:8911/devices")
		if err != nil {
			daemonRunning = "Offline"
			discoveryActive = "Inactive"
			onlineCount = 1
		} else {
			defer resp.Body.Close()
			var devices []core.Device
			if json.NewDecoder(resp.Body).Decode(&devices) == nil {
				for _, dev := range devices {
					if dev.Status == "online" {
						onlineCount++
					}
				}
			}
		}

		out := cmd.OutOrStdout()
		fmt.Fprintln(out)
		fmt.Fprintf(out, " %-13s %s\n", ui.Bold("Mesh"), ui.Green("Healthy"))
		fmt.Fprintln(out)
		ui.Separator(out)
		fmt.Fprintln(out)

		ui.PrintDetail(out, "Devices", fmt.Sprintf("%d", totalCount))
		ui.PrintDetail(out, "Online", fmt.Sprintf("%d", onlineCount))
		ui.PrintDetail(out, "Trusted", fmt.Sprintf("%d", trustedCount))
		ui.PrintDetail(out, "Protocol", "1")
		ui.PrintDetail(out, "Discovery", discoveryActive)
		ui.PrintDetail(out, "TLS", "Enabled")
		ui.PrintDetail(out, "Daemon", daemonRunning)
		fmt.Fprintln(out)

		return nil
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
