package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
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

		if err := ensureDaemonRunning(); err != nil {
			return err
		}

		cfg, err := core.LoadConfig()
		if err != nil {
			return err
		}

		// Fetch trusted devices list from daemon
		resp, err := localClient().Get("https://127.0.0.1:8911/devices")
		if err != nil {
			return fmt.Errorf("failed to fetch devices: %w", err)
		}
		defer resp.Body.Close()

		var devices []core.Device
		if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
			return fmt.Errorf("malformed daemon payload: %w", err)
		}

		commandStr := strings.Join(args, " ")

		// Build target devices list (local + remote trusted ones)
		type TargetHost struct {
			ID   string
			Name string
			IP   string
			OS   string
		}

		targets := []TargetHost{
			{
				ID:   cfg.DeviceID,
				Name: cfg.DeviceName + "*",
				IP:   "127.0.0.1",
				OS:   "local",
			},
		}

		for _, dev := range devices {
			targets = append(targets, TargetHost{
				ID:   dev.ID,
				Name: dev.Name,
				IP:   dev.IP,
				OS:   dev.OS,
			})
		}

		// Start execution timer
		startTime := time.Now()

		type HostResult struct {
			Outputs []string
			Failed  bool
			ErrMsg  string
		}

		results := make(map[string]*HostResult)
		var resultsMu sync.Mutex
		var wg sync.WaitGroup

		for _, target := range targets {
			wg.Add(1)
			go func(host TargetHost) {
				defer wg.Done()

				hr := &HostResult{Outputs: []string{}}
				defer func() {
					resultsMu.Lock()
					results[host.ID] = hr
					resultsMu.Unlock()
				}()

				// Call local daemon run endpoint which forwards requests over mTLS to target IP
				runURL := fmt.Sprintf("https://127.0.0.1:8911/run?ip=%s&command=%s", host.IP, url.QueryEscape(commandStr))
				rResp, rErr := localClient().Post(runURL, "application/json", nil)
				if rErr != nil {
					hr.Failed = true
					hr.ErrMsg = rErr.Error()
					return
				}
				defer rResp.Body.Close()

				if rResp.StatusCode != http.StatusOK {
					hr.Failed = true
					body, _ := io.ReadAll(rResp.Body)
					hr.ErrMsg = string(body)
					if hr.ErrMsg == "" {
						hr.ErrMsg = fmt.Sprintf("HTTP %d", rResp.StatusCode)
					}
					return
				}

				decoder := json.NewDecoder(rResp.Body)
				for {
					var out core.JobOutput
					if decodeErr := decoder.Decode(&out); decodeErr != nil {
						if decodeErr == io.EOF {
							break
						}
						hr.Failed = true
						hr.ErrMsg = decodeErr.Error()
						break
					}

					switch out.Type {
					case "stdout":
						hr.Outputs = append(hr.Outputs, out.Content)
					case "stderr":
						hr.Outputs = append(hr.Outputs, ui.Red(out.Content))
					case "error":
						hr.Failed = true
						hr.ErrMsg = out.Content
					case "exit":
						if out.ExitCode != 0 {
							hr.Failed = true
							hr.ErrMsg = fmt.Sprintf("exit code %d", out.ExitCode)
						}
					}
				}
			}(target)
		}

		wg.Wait()
		elapsed := time.Since(startTime).Truncate(time.Millisecond)

		// Print clean typography Design System output report
		fmt.Fprintln(cmd.OutOrStdout())
		ui.Header(cmd.OutOrStdout(), "Running")
		fmt.Fprintf(cmd.OutOrStdout(), "\n  %s\n\n", ui.Bold(commandStr))
		ui.Separator(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout())

		successCount := 0
		failureCount := 0

		for _, target := range targets {
			res := results[target.ID]
			symbol := ui.Success()
			if res.Failed {
				symbol = ui.ErrorSymbol()
				failureCount++
			} else {
				successCount++
			}

			fmt.Fprintf(cmd.OutOrStdout(), " %s %s\n\n", symbol, ui.Bold(target.Name))
			for _, line := range res.Outputs {
				fmt.Fprintf(cmd.OutOrStdout(), "    %s\n", line)
			}
			if res.Failed && res.ErrMsg != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "    %s\n", ui.Red("Error: "+res.ErrMsg))
			}
			fmt.Fprintln(cmd.OutOrStdout())
		}

		ui.Separator(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout())

		ui.Header(cmd.OutOrStdout(), "Completed")
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintf(cmd.OutOrStdout(), "    %d success\n", successCount)
		if failureCount > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "    %d failed\n", failureCount)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "    %s\n\n", elapsed)

		if failureCount > 0 {
			return errors.New("command execution failed on some hosts")
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
}
