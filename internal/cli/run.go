package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/orion-infra/orion/internal/core"
	"github.com/orion-infra/orion/internal/ui"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:                "run <command>",
	Short:              "Execute a command across all paired devices in parallel",
	Example:            "  $ orion run uname -a",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
			return cmd.Help()
		}

		engine, cfg, err := getEngine()
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("orion is not initialized. Run %s first", ui.Bold("orion init"))
			}
			return err
		}

		commandStr := strings.Join(args, " ")

		devices, err := engine.GetDevices()
		if err != nil {
			return err
		}

		// Build target devices list (local + remote)
		targets := []core.Device{
			{
				ID:     cfg.DeviceID,
				Name:   cfg.DeviceName,
				OS:     runtime.GOOS,
				Status: "online",
			},
		}
		targets = append(targets, devices...)

		// Header Renders
		ui.Header(cmd.OutOrStdout(), "Running")
		fmt.Fprintf(cmd.OutOrStdout(), "\n  %s\n\n", ui.Bold(commandStr))
		ui.Separator(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Start execution timer
		startTime := time.Now()
		ch := engine.RunCommand(ctx, targets, commandStr)

		// Accumulate output lines per device in memory
		outputs := make(map[string][]string)
		deviceFailed := make(map[string]bool)
		deviceSimulated := make(map[string]bool)

		for line := range ch {
			if line.Err != nil {
				deviceFailed[line.DeviceName] = true
				outputs[line.DeviceName] = append(outputs[line.DeviceName], ui.Red("Error: "+line.Err.Error()))
			} else {
				if line.IsStderr {
					// Check if this is the simulation notice
					if strings.Contains(line.Content, "Execution engine unavailable") ||
						strings.Contains(line.Content, "Simulation") ||
						strings.Contains(line.Content, "Learn more") {
						deviceSimulated[line.DeviceName] = true
						outputs[line.DeviceName] = append(outputs[line.DeviceName], ui.Gray(line.Content))
					} else {
						outputs[line.DeviceName] = append(outputs[line.DeviceName], ui.Red(line.Content))
					}
				} else {
					outputs[line.DeviceName] = append(outputs[line.DeviceName], line.Content)
				}
			}
		}

		elapsed := time.Since(startTime).Truncate(time.Millisecond)

		successCount := 0
		simulatedCount := 0

		// Print outputs per device
		for _, d := range targets {
			var symbol string
			var label string

			if d.ID == cfg.DeviceID {
				label = d.Name + "*"
				if deviceFailed[d.Name] {
					symbol = ui.ErrorSymbol()
				} else {
					symbol = ui.Success()
					successCount++
				}
			} else {
				label = d.Name
				if deviceFailed[d.Name] {
					symbol = ui.ErrorSymbol()
				} else {
					symbol = ui.Warning()
					simulatedCount++
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), " %s %s\n\n", symbol, ui.Bold(label))
			for _, line := range outputs[d.Name] {
				fmt.Fprintf(cmd.OutOrStdout(), "    %s\n", line)
			}
			fmt.Fprintln(cmd.OutOrStdout())
		}

		ui.Separator(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout())

		// Report completion metric
		ui.Header(cmd.OutOrStdout(), "Completed")
		fmt.Fprintln(cmd.OutOrStdout())

		fmt.Fprintf(cmd.OutOrStdout(), "    %d success\n", successCount)
		if simulatedCount > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "    %d simulated\n", simulatedCount)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "    %s\n", elapsed)

		if deviceFailed[cfg.DeviceName] {
			return fmt.Errorf("local execution failed")
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
}
