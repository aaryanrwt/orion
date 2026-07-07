package cli

import (
	"fmt"
	"runtime"

	"github.com/orion-infra/orion/internal/ui"
	"github.com/spf13/cobra"
)

var (
	// Version is the current version of the Orion CLI.
	Version = "v0.1.0-beta"
	// Commit is the git commit hash injected during build.
	Commit = "none"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Display Orion CLI version details",
	Example: "  $ orion version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), ui.BrandHeader(Version))
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "  Run commands across computers you own.")
		fmt.Fprintln(cmd.OutOrStdout())
		ui.Separator(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout())

		ui.Header(cmd.OutOrStdout(), "Metadata")
		fmt.Fprintln(cmd.OutOrStdout())

		ui.PrintDetail(cmd.OutOrStdout(), "Version", Version)
		if Commit != "none" && Commit != "dev" && Commit != "" {
			ui.PrintDetail(cmd.OutOrStdout(), "Commit", Commit)
		}
		ui.PrintDetail(cmd.OutOrStdout(), "Platform", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH))
		return nil
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
