package cli

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/orion-infra/orion/internal/core"
	"github.com/orion-infra/orion/internal/ui"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "orion",
	Short: "Run one command across every computer you own.",
	Long: ui.Bold("Orion") + " - Run one command across every computer you own.\n\n" +
		"A minimal, secure, and fast developer utility for parallel command execution.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "daemon" || cmd.Name() == "init" || cmd.Name() == "respond" {
			return nil
		}

		// Check if config exists
		_, err := core.LoadConfig()
		if err != nil {
			return nil // Let subcommand handle uninitialized state
		}

		// Check if daemon is active and has pending pairing requests
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: 80 * time.Millisecond,
		}

		resp, err := client.Get("https://127.0.0.1:8911/pending")
		if err == nil {
			defer resp.Body.Close()
			var status map[string]interface{}
			if json.NewDecoder(resp.Body).Decode(&status) == nil {
				if pending, _ := status["pending"].(bool); pending {
					deviceName, _ := status["device_name"].(string)
					deviceID, _ := status["device_id"].(string)
					fingerprint, _ := status["fingerprint"].(string)
					_ = HandleConsentPrompt(deviceName, deviceID, fingerprint)
				}
			}
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		summary, reason, try := translateError(err)
		ui.PrintError(os.Stderr, summary, reason, try)
		os.Exit(1)
	}
}

func translateError(err error) (string, string, string) {
	msg := err.Error()
	if strings.Contains(msg, "accepts 1 arg") {
		return "No device specified", "The join command requires a target device ID.", "orion join <device-id>"
	}
	if strings.Contains(msg, "unknown command") {
		return "Unknown command", msg, "orion help"
	}
	return "Command failed", msg, ""
}

func init() {
	RootCmd.CompletionOptions.DisableDefaultCmd = true
	RootCmd.SetHelpTemplate(helpTemplate())
}

func helpTemplate() string {
	devicesCount := "Uninitialized"
	ready := "No"
	tipHeader := "Initialize Orion"
	tipCmd := "orion init"

	cfg, err := core.LoadConfig()
	if err == nil && cfg != nil && cfg.DeviceID != "" {
		ready = "Yes"
		devicesCount = fmt.Sprintf("%d connected", len(cfg.Devices))
		if len(cfg.Devices) == 0 {
			tipHeader = "Pair another computer"
			tipCmd = "orion connect"
		} else {
			tipHeader = "Run a command"
			tipCmd = "orion run uname -a"
		}
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(ui.BrandHeader(Version) + "\n\n")
	sb.WriteString("  Run commands across computers you own.\n\n")

	// Quick Start Section
	sb.WriteString(ui.Gray("──────────────────────────────────────────────────────────") + "\n\n")
	sb.WriteString(" " + ui.Bold("Quick Start") + "\n\n")
	sb.WriteString("   " + ui.Blue("init") + "        Initialize Orion\n")
	sb.WriteString("   " + ui.Blue("connect") + "     Pair another computer\n")
	sb.WriteString("   " + ui.Blue("run") + "         Execute a command\n")
	sb.WriteString("   " + ui.Blue("doctor") + "      Diagnose problems\n\n")

	// Status Section
	sb.WriteString(ui.Gray("──────────────────────────────────────────────────────────") + "\n\n")
	sb.WriteString(" " + ui.Bold("Status") + "\n\n")
	sb.WriteString(fmt.Sprintf("   %-11s %s\n", "Devices", devicesCount))
	sb.WriteString(fmt.Sprintf("   %-11s %s\n\n", "Ready", ready))

	// Tip Section
	sb.WriteString(ui.Gray("──────────────────────────────────────────────────────────") + "\n\n")
	sb.WriteString(" " + ui.Bold("Tip") + "\n\n")
	sb.WriteString("   " + tipHeader + "\n\n")
	sb.WriteString("   " + ui.Blue(tipCmd) + "\n")

	return sb.String()
}

// getEngine loads local config and returns a new Engine implementation.
func getEngine() (core.Engine, *core.Config, error) {
	cfg, err := core.LoadConfig()
	if err != nil {
		return nil, nil, err
	}
	return core.NewMockEngine(cfg), cfg, nil
}
