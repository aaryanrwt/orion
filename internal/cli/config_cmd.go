package cli

import (
	"fmt"

	"github.com/orion-infra/orion/internal/core"
	"github.com/orion-infra/orion/internal/ui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure Orion local preferences",
}

var configNameCmd = &cobra.Command{
	Use:   "name <new-name>",
	Short: "Rename this machine's local display name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		newName := args[0]
		if err := core.RenameLocalDevice(newName); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\n  %s Local device renamed to \"%s\"\n\n", ui.Success(), newName)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configNameCmd)
	RootCmd.AddCommand(configCmd)
}
