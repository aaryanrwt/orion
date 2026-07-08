package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/orion-infra/orion/internal/core"
	"github.com/orion-infra/orion/internal/ui"
	"github.com/spf13/cobra"
)

var hardwareCmd = &cobra.Command{
	Use:   "hardware",
	Short: "Display hardware specs and CPU/GPU details of all mesh machines",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureDaemonRunning(); err != nil {
			return err
		}

		cfg, err := core.LoadConfig()
		if err != nil {
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

		localHW := core.DetectHardware()

		fmt.Fprintln(os.Stdout)
		ui.Header(os.Stdout, "Mesh Hardware")
		fmt.Fprintln(os.Stdout)

		// Print local machine
		fmt.Fprintf(os.Stdout, "   %s (Local)\n", ui.Bold(cfg.DeviceName+"*"))
		fmt.Fprintf(os.Stdout, "     %-15s %s\n", "CPU:", localHW.CPU)
		fmt.Fprintf(os.Stdout, "     %-15s %s\n", "RAM:", localHW.RAM)
		fmt.Fprintf(os.Stdout, "     %-15s %s\n", "GPU:", localHW.GPU)
		fmt.Fprintf(os.Stdout, "     %-15s %s\n", "VRAM:", localHW.VRAM)
		fmt.Fprintf(os.Stdout, "     %-15s %s\n", "CUDA:", localHW.CUDA)
		fmt.Fprintf(os.Stdout, "     %-15s %s\n", "Driver:", localHW.DriverVersion)
		fmt.Fprintf(os.Stdout, "     %-15s %s\n", "CUDA Version:", localHW.CUDAVersion)
		fmt.Fprintf(os.Stdout, "     %-15s %s\n", "OS:", runtime.GOOS)
		fmt.Fprintf(os.Stdout, "     %-15s %s\n", "Arch:", runtime.GOARCH)
		fmt.Fprintln(os.Stdout)

		// Print peers
		for _, dev := range devices {
			statusSymbol := " "
			if dev.Status == "online" {
				statusSymbol = "✓"
			}
			displayName := dev.Name
			if dev.Alias != "" {
				displayName = dev.Alias
			}
			fmt.Fprintf(os.Stdout, "   %s %s (%s)\n", statusSymbol, ui.Bold(displayName), dev.ID)
			if dev.Status == "online" {
				fmt.Fprintf(os.Stdout, "     %-15s %s\n", "CPU:", dev.Hardware.CPU)
				fmt.Fprintf(os.Stdout, "     %-15s %s\n", "RAM:", dev.Hardware.RAM)
				fmt.Fprintf(os.Stdout, "     %-15s %s\n", "GPU:", dev.Hardware.GPU)
				fmt.Fprintf(os.Stdout, "     %-15s %s\n", "VRAM:", dev.Hardware.VRAM)
				fmt.Fprintf(os.Stdout, "     %-15s %s\n", "CUDA:", dev.Hardware.CUDA)
				fmt.Fprintf(os.Stdout, "     %-15s %s\n", "Driver:", dev.Hardware.DriverVersion)
				fmt.Fprintf(os.Stdout, "     %-15s %s\n", "CUDA Version:", dev.Hardware.CUDAVersion)
				fmt.Fprintf(os.Stdout, "     %-15s %s\n", "OS:", dev.OS)
				fmt.Fprintf(os.Stdout, "     %-15s %s\n", "Arch:", dev.Hardware.Arch)
			} else {
				fmt.Fprintf(os.Stdout, "     %s\n", ui.Gray("Offline"))
			}
			fmt.Fprintln(os.Stdout)
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(hardwareCmd)
}
