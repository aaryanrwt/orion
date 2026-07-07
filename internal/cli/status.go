package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/orion-infra/orion/internal/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Display a high-level summary of your Orion network",
	Example: "  $ orion status",
	RunE: func(cmd *cobra.Command, args []string) error {
		engine, _, err := getEngine()
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("orion is not initialized. Run %s first", ui.Bold("orion init"))
			}
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout())
		ui.Header(cmd.OutOrStdout(), "Orion Network Status")
		fmt.Fprintln(cmd.OutOrStdout())
		ui.Separator(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout())

		devices, err := engine.GetDevices()
		if err != nil {
			return err
		}

		// Count devices (local + paired online devices)
		onlineCount := 1 // Local device is always online
		for _, dev := range devices {
			if dev.Status == "online" {
				onlineCount++
			}
		}

		// Total paired devices + local
		totalCount := len(devices) + 1

		connectedVal := fmt.Sprintf("%d devices (%d online)", totalCount, onlineCount)
		if len(devices) == 0 {
			connectedVal = "1 device (local)"
		}

		ui.Header(cmd.OutOrStdout(), "Status")
		fmt.Fprintln(cmd.OutOrStdout())
		ui.PrintDetail(cmd.OutOrStdout(), "Connected", connectedVal)
		ui.PrintDetail(cmd.OutOrStdout(), "Ready", "Yes")
		fmt.Fprintln(cmd.OutOrStdout())

		ui.Separator(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout())

		ui.Header(cmd.OutOrStdout(), "Tip")
		fmt.Fprintln(cmd.OutOrStdout())
		if len(devices) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "   Pair another computer")
			fmt.Fprintln(cmd.OutOrStdout())
			ui.Suggestion(cmd.OutOrStdout(), "orion join <device-id>")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "   Run a command")
			fmt.Fprintln(cmd.OutOrStdout())
			ui.Suggestion(cmd.OutOrStdout(), "orion run uname -a")
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
