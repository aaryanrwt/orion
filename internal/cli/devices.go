package cli

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/orion-infra/orion/internal/ui"
	"github.com/spf13/cobra"
)

var devicesCmd = &cobra.Command{
	Use:     "devices",
	Short:   "List all paired devices and their statuses",
	Example: "  $ orion devices",
	Aliases: []string{"list", "hosts"},
	RunE: func(cmd *cobra.Command, args []string) error {
		engine, cfg, err := getEngine()
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("orion is not initialized. Run %s first", ui.Bold("orion init"))
			}
			return err
		}

		ui.Header(cmd.OutOrStdout(), "Orion Mesh Devices")
		fmt.Fprintln(cmd.OutOrStdout())

		// We add "  " to the first column header to indent the table grid
		headers := []string{"  DEVICE ID", "NAME", "OS", "STATUS", "LATENCY"}
		var rows [][]string

		// Add local machine as initiator
		rows = append(rows, []string{
			"  " + cfg.DeviceID,
			cfg.DeviceName + "*",
			runtime.GOOS,
			ui.Green("online"),
			"-",
		})

		devices, err := engine.GetDevices()
		if err != nil {
			return err
		}

		// Add paired machines
		for _, dev := range devices {
			statusColor := ui.Green(dev.Status)
			if dev.Status == "offline" {
				statusColor = ui.Red(dev.Status)
			}
			rows = append(rows, []string{
				"  " + dev.ID,
				dev.Name,
				dev.OS,
				statusColor,
				dev.Latency,
			})
		}

		ui.PrintTable(cmd.OutOrStdout(), headers, rows)

		if len(devices) == 0 {
			ui.Suggestion(cmd.OutOrStdout(), "orion join <device-id>")
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(devicesCmd)
}
