package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var respondCmd = &cobra.Command{
	Use:     "respond",
	Short:   "Respond to pending incoming connection pairing requests",
	Example: "  $ orion respond",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureDaemonRunning(); err != nil {
			return err
		}

		// Query pending pairing requests
		resp, err := localClient().Get("https://127.0.0.1:8911/pending")
		if err != nil {
			return fmt.Errorf("failed to query pending request: %w", err)
		}
		defer resp.Body.Close()

		var status map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			return fmt.Errorf("malformed daemon payload: %w", err)
		}

		pending, _ := status["pending"].(bool)
		if !pending {
			fmt.Fprintln(os.Stdout, "\n  No pending connection requests.")
			return nil
		}

		deviceName, _ := status["device_name"].(string)
		deviceID, _ := status["device_id"].(string)
		fingerprint, _ := status["fingerprint"].(string)

		return HandleConsentPrompt(deviceName, deviceID, fingerprint)
	},
}

func init() {
	RootCmd.AddCommand(respondCmd)
}
