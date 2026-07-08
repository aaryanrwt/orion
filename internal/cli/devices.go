package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/orion-infra/orion/internal/core"
	"github.com/orion-infra/orion/internal/ui"
	"github.com/spf13/cobra"
)

var devicesVerbose bool

var devicesCmd = &cobra.Command{
	Use:     "devices",
	Short:   "List all paired devices and their statuses",
	Example: "  $ orion devices",
	Aliases: []string{"list", "hosts"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureDaemonRunning(); err != nil {
			return err
		}

		resp, err := localClient().Get("https://127.0.0.1:8911/devices")
		if err != nil {
			return fmt.Errorf("failed to fetch devices: %w", err)
		}
		defer resp.Body.Close()

		var devices []core.Device
		if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
			return fmt.Errorf("malformed daemon payload: %w", err)
		}

		cfg, err := core.LoadConfig()
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("orion is not initialized. Run %s first", ui.Bold("orion init"))
			}
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout())
		ui.Header(cmd.OutOrStdout(), "Devices")
		fmt.Fprintln(cmd.OutOrStdout())

		if !devicesVerbose {
			// Minimal Layout
			// Print local machine
			fmt.Fprintf(cmd.OutOrStdout(), "   ✓ %s\n", ui.Bold(cfg.DeviceName+"*"))
			fmt.Fprintf(cmd.OutOrStdout(), "     %s\n", cfg.DeviceID)
			fmt.Fprintf(cmd.OutOrStdout(), "     %s\n\n", runtime.GOOS)

			// Print peers
			for _, dev := range devices {
				statusSymbol := "✓"
				statusText := dev.OS
				if dev.Status == "offline" {
					statusSymbol = "⚪"
					statusText = "Offline"
					if dev.Latency != "-" && strings.Contains(dev.Latency, "Last seen") {
						statusText = "Offline (" + dev.Latency + ")"
					}
				}
				displayName := dev.Name
				if dev.Alias != "" {
					displayName = dev.Alias
				}
				fmt.Fprintf(cmd.OutOrStdout(), "   %s %s\n", statusSymbol, ui.Bold(displayName))
				fmt.Fprintf(cmd.OutOrStdout(), "     %s\n", dev.ID)
				fmt.Fprintf(cmd.OutOrStdout(), "     %s\n\n", statusText)
			}
		} else {
			// Verbose Layout
			// Print local machine
			localStatus := ui.Success()
			fmt.Fprintf(cmd.OutOrStdout(), "   %s %s (Local)\n", localStatus, ui.Bold(cfg.DeviceName+"*"))
			fmt.Fprintf(cmd.OutOrStdout(), "     %-15s %s\n", "ID:", cfg.DeviceID)
			fmt.Fprintf(cmd.OutOrStdout(), "     %-15s %s\n", "OS:", runtime.GOOS)
			fmt.Fprintf(cmd.OutOrStdout(), "     %-15s %s\n", "Fingerprint:", cfg.PublicKey)
			if cfg.CertificatePEM != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "     %-15s %s\n", "Certificate:", "[PEM Block]")
			}
			fmt.Fprintln(cmd.OutOrStdout())

			// Print peers
			for _, dev := range devices {
				statusSymbol := ui.Success()
				if dev.Status == "offline" {
					statusSymbol = ui.ErrorSymbol()
				}
				displayName := dev.Name
				if dev.Alias != "" {
					displayName = fmt.Sprintf("%s (%s)", dev.Name, dev.Alias)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "   %s %s\n", statusSymbol, ui.Bold(displayName))
				fmt.Fprintf(cmd.OutOrStdout(), "     %-15s %s\n", "ID:", dev.ID)
				fmt.Fprintf(cmd.OutOrStdout(), "     %-15s %s\n", "OS:", dev.OS)
				fmt.Fprintf(cmd.OutOrStdout(), "     %-15s %s\n", "Fingerprint:", dev.Fingerprint)
				fmt.Fprintf(cmd.OutOrStdout(), "     %-15s %s\n", "Latency:", dev.Latency)
				if dev.TrustTimestamp != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "     %-15s %s\n", "Trust Time:", dev.TrustTimestamp)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "     %-15s %d\n", "Protocol:", dev.Protocol)
				fmt.Fprintf(cmd.OutOrStdout(), "     %-15s %s\n", "Version:", dev.Version)
				fmt.Fprintf(cmd.OutOrStdout(), "     %-15s %s\n", "Status:", dev.Status)
				if dev.Certificate != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "     %-15s %s\n", "Certificate:", "[PEM Block]")
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}
		}

		if len(devices) == 0 {
			ui.Separator(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout())
			ui.Header(cmd.OutOrStdout(), "Tip")
			fmt.Fprintln(cmd.OutOrStdout())
			ui.Suggestion(cmd.OutOrStdout(), "orion connect")
		}

		return nil
	},
}

var renameCmd = &cobra.Command{
	Use:     "rename <device-id> <new-alias>",
	Short:   "Rename a trusted peer device display alias",
	Example: "  $ orion rename ORN-9AF2-81D7 \"Home PC\"",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureDaemonRunning(); err != nil {
			return err
		}

		id := args[0]
		alias := args[1]

		resp, err := localClient().Post(fmt.Sprintf("https://127.0.0.1:8911/devices/rename?id=%s&alias=%s", id, url.QueryEscape(alias)), "application/json", nil)
		if err != nil {
			return fmt.Errorf("failed to rename device: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("device %s not found in trusted list", id)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "\n  %s Device %s alias renamed to \"%s\"\n\n", ui.Success(), id, alias)
		return nil
	},
}

var removeCmd = &cobra.Command{
	Use:     "remove <device-id>",
	Short:   "Revoke trust and remove a paired remote device",
	Example: "  $ orion remove ORN-9AF2-81D7",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureDaemonRunning(); err != nil {
			return err
		}

		id := args[0]

		resp, err := localClient().Post(fmt.Sprintf("https://127.0.0.1:8911/devices/remove?id=%s", id), "application/json", nil)
		if err != nil {
			return fmt.Errorf("failed to remove device: %w", err)
		}
		defer resp.Body.Close()

		fmt.Fprintf(cmd.OutOrStdout(), "\n  %s Revoked trust and removed device %s\n\n", ui.Success(), id)
		return nil
	},
}

func init() {
	devicesCmd.Flags().BoolVarP(&devicesVerbose, "verbose", "v", false, "Show detailed device metadata")
	RootCmd.AddCommand(devicesCmd)
	RootCmd.AddCommand(renameCmd)
	RootCmd.AddCommand(removeCmd)
}
