package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/orion-infra/orion/internal/core"
	"github.com/orion-infra/orion/internal/ui"
	"github.com/spf13/cobra"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Display all available Ollama models across the mesh",
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
			return fmt.Errorf("failed to fetch devices models list: %w", err)
		}
		defer resp.Body.Close()

		var devices []core.Device
		if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
			return fmt.Errorf("malformed daemon payload: %w", err)
		}

		// Fetch local models
		localOllama, localModels := coreGetLocalOllamaModels()

		fmt.Fprintln(os.Stdout)
		ui.Header(os.Stdout, "Available Models")
		fmt.Fprintln(os.Stdout)

		// Print local machine
		fmt.Fprintf(os.Stdout, "   %s (Local)\n", ui.Bold(cfg.DeviceName+"*"))
		if localOllama {
			if len(localModels) == 0 {
				fmt.Fprintln(os.Stdout, "     No models downloaded")
			}
			for _, m := range localModels {
				sizeGB := float64(m.Size) / 1024 / 1024 / 1024
				fmt.Fprintf(os.Stdout, "     ✓ %-28s %-10s %-12s %s\n", m.Name, fmt.Sprintf("%.1f GB", sizeGB), m.Quantization, m.Status)
			}
		} else {
			fmt.Fprintln(os.Stdout, "     Ollama not installed or running")
		}
		fmt.Fprintln(os.Stdout)

		// Print peers
		for _, dev := range devices {
			if dev.Status == "online" {
				displayName := dev.Name
				if dev.Alias != "" {
					displayName = dev.Alias
				}
				fmt.Fprintf(os.Stdout, "   %s (%s)\n", ui.Bold(displayName), dev.ID)
				if dev.OllamaInstalled {
					if len(dev.Models) == 0 {
						fmt.Fprintln(os.Stdout, "     No models downloaded")
					}
					for _, m := range dev.Models {
						sizeGB := float64(m.Size) / 1024 / 1024 / 1024
						fmt.Fprintf(os.Stdout, "     ✓ %-28s %-10s %-12s %s\n", m.Name, fmt.Sprintf("%.1f GB", sizeGB), m.Quantization, m.Status)
					}
				} else {
					fmt.Fprintln(os.Stdout, "     Ollama not installed or running")
				}
				fmt.Fprintln(os.Stdout)
			}
		}

		return nil
	},
}

func coreGetLocalOllamaModels() (bool, []core.OllamaModel) {
	client := &http.Client{Timeout: 300 * time.Millisecond}
	resp, err := client.Get("http://127.0.0.1:11434/api/tags")
	if err != nil {
		return false, nil
	}
	defer resp.Body.Close()

	// Get running models from /api/ps
	runningModels := make(map[string]bool)
	psResp, psErr := client.Get("http://127.0.0.1:11434/api/ps")
	if psErr == nil {
		defer psResp.Body.Close()
		var psPayload struct {
			Models []struct {
				Name string `json:"name"`
			} `json:"models"`
		}
		if json.NewDecoder(psResp.Body).Decode(&psPayload) == nil {
			for _, m := range psPayload.Models {
				runningModels[m.Name] = true
			}
		}
	}

	var payload struct {
		Models []struct {
			Name    string `json:"name"`
			Size    int64  `json:"size"`
			Details struct {
				Quantization string `json:"quantization_level"`
			} `json:"details"`
		} `json:"models"`
	}
	if json.NewDecoder(resp.Body).Decode(&payload) != nil {
		return true, nil
	}

	list := make([]core.OllamaModel, len(payload.Models))
	for i, m := range payload.Models {
		status := "Idle"
		if runningModels[m.Name] {
			status = "Running"
		}
		quant := m.Details.Quantization
		if quant == "" {
			quant = "-"
		}
		list[i] = core.OllamaModel{
			Name:         m.Name,
			Size:         m.Size,
			Quantization: quant,
			Status:       status,
		}
	}
	return true, list
}

func init() {
	RootCmd.AddCommand(modelsCmd)
}
