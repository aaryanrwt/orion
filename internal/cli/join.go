package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/orion-infra/orion/internal/core"
	"github.com/orion-infra/orion/internal/ui"
	"github.com/spf13/cobra"
)

var joinCmd = &cobra.Command{
	Use:     "join <device-id>",
	Short:   "Securely pair and trust a new remote device",
	Example: "  $ orion join vega-88e2-9b2f",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetID := args[0]

		ui.Header(cmd.OutOrStdout(), "Join Device")
		fmt.Fprintln(cmd.OutOrStdout())

		// Basic validation of Device ID format
		if !strings.Contains(targetID, "-") {
			return fmt.Errorf("invalid device ID format. Device ID must be of the form <hostname>-<suffix>")
		}

		engine, cfg, err := getEngine()
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("orion is not initialized. Run %s first", ui.Bold("orion init"))
			}
			return err
		}

		if cfg.DeviceID == targetID {
			return fmt.Errorf("cannot join a device to itself")
		}

		// Check if already joined
		for _, dev := range cfg.Devices {
			if dev.ID == targetID {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s Device is already paired with Orion.\n", ui.Warning())
				return nil
			}
		}

		fmt.Fprintln(cmd.OutOrStdout(), "  Searching nearby systems...")
		ui.SimulateSpinner(cmd.OutOrStdout(), "Establishing secure connection", 1200*time.Millisecond)

		// Parse host name from target device ID
		parts := strings.Split(targetID, "-")
		deviceName := parts[0]
		if deviceName == "" {
			deviceName = "remote-device"
		}

		// Detect OS based on name or randomize
		targetOS := "linux"
		if strings.Contains(deviceName, "mac") || strings.Contains(deviceName, "apple") || strings.Contains(deviceName, "vega") {
			targetOS = "darwin"
		} else if strings.Contains(deviceName, "win") {
			targetOS = "windows"
		}

		newDevice := core.Device{
			ID:      targetID,
			Name:    deviceName,
			OS:      targetOS,
			IP:      "192.168.1.15",
			Latency: "4.5ms",
			Status:  "online",
		}

		if err := engine.AddDevice(newDevice); err != nil {
			return fmt.Errorf("failed to save paired device info: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "  %s Secure connection established with %s\n", ui.Success(), targetID)
		fmt.Fprintf(cmd.OutOrStdout(), "  %s Added device \"%s\"\n", ui.Success(), deviceName)

		ui.Suggestion(cmd.OutOrStdout(), "orion run uname -a")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(joinCmd)
}
