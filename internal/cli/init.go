package cli

import (
	"fmt"

	"github.com/orion-infra/orion/internal/core"
	"github.com/orion-infra/orion/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "Initialize local machine credentials and configuration",
	Example: "  $ orion init",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Header(cmd.OutOrStdout(), "Initialize Orion")
		fmt.Fprintln(cmd.OutOrStdout())

		// Check if already initialized
		cfg, err := core.LoadConfig()
		if err == nil && cfg != nil && cfg.DeviceID != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s Orion is already initialized on this machine.\n", ui.Warning())
			ui.PrintDetail(cmd.OutOrStdout(), "Device ID", cfg.DeviceID)
			ui.Suggestion(cmd.OutOrStdout(), "orion status")
			return nil
		}

		// Initialize
		cfg, err = core.InitializeLocalDevice()
		if err != nil {
			return fmt.Errorf("initialization failed: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "  %s Orion initialized successfully.\n", ui.Success())
		ui.PrintDetail(cmd.OutOrStdout(), "Device ID", cfg.DeviceID)
		ui.Suggestion(cmd.OutOrStdout(), "orion join <device-id>")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
