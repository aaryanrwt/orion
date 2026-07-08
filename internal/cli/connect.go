package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/orion-infra/orion/internal/core"
	"github.com/orion-infra/orion/internal/ui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var connectCmd = &cobra.Command{
	Use:     "connect",
	Short:   "Discover nearby Orion devices and pair with them",
	Example: "  $ orion connect",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Ensure config exists
		_, err := core.LoadConfig()
		if err != nil {
			return fmt.Errorf("orion is not initialized. Run %s first", ui.Bold("orion init"))
		}

		// 2. Ensure background daemon is running
		if err := ensureDaemonRunning(); err != nil {
			return err
		}

		// 3. Interactive connection loop
		return runInteractiveConnect(cmd)
	},
}

func init() {
	RootCmd.AddCommand(connectCmd)
}

type menuOption struct {
	label       string
	id          string
	isPeer      bool
	peerIP      string
	peerOS      string
	peerStatus  string
	peerLatency string
}

func runInteractiveConnect(cmd *cobra.Command) error {
	// Put terminal in raw mode for arrow key selection
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("failed to initialize interactive terminal: %w", err)
	}
	defer term.Restore(fd, oldState)

	// ANSI clear screen helper for interactive redraws
	clearLines := func(n int) {
		if n > 0 {
			fmt.Printf("\033[%dA\033[J", n)
		}
	}

	selectedIdx := 0
	linesDrawn := 0
	peersList := []core.DiscoveredPeer{}
	var lastRefresh time.Time

	// Local helper to fetch discovered devices from daemon
	refreshPeers := func() {
		if time.Since(lastRefresh) < 400*time.Millisecond {
			return
		}
		lastRefresh = time.Now()
		resp, err := localClient().Get("https://127.0.0.1:8911/nearby")
		if err == nil {
			defer resp.Body.Close()
			var list []core.DiscoveredPeer
			if json.NewDecoder(resp.Body).Decode(&list) == nil {
				peersList = list
			}
		}
	}

	// Read key strokes in background
	keyCh := make(chan string)
	go func() {
		var buf [3]byte
		for {
			n, err := os.Stdin.Read(buf[:])
			if err != nil {
				return
			}
			if n == 1 {
				if buf[0] == 3 { // Ctrl+C
					keyCh <- "ctrlc"
					return
				}
				if buf[0] == 13 || buf[0] == 10 {
					keyCh <- "enter"
					continue
				}
				keyCh <- string(buf[:1])
			}
			if n == 3 && buf[0] == 27 && buf[1] == 91 {
				if buf[2] == 65 {
					keyCh <- "up"
				}
				if buf[2] == 66 {
					keyCh <- "down"
				}
			}
		}
	}()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	// Initial fetch
	refreshPeers()

	for {
		// Compile menu choices
		options := []menuOption{}

		// Load config to check trusted status
		cfg, _ := core.LoadConfig()
		trustedMap := make(map[string]bool)
		for _, dev := range cfg.Devices {
			trustedMap[dev.ID] = true
		}

		for _, p := range peersList {
			statusLabel := "New"
			if trustedMap[p.ID] {
				statusLabel = "Trusted"
			}
			options = append(options, menuOption{
				label:       p.Name,
				id:          p.ID,
				isPeer:      true,
				peerIP:      p.IP,
				peerOS:      p.OS,
				peerStatus:  statusLabel,
				peerLatency: p.Latency,
			})
		}

		// Fallback items
		options = append(options, menuOption{label: "Manual Device ID", id: "manual_id"})
		options = append(options, menuOption{label: "Manual IP Address", id: "manual_ip"})

		// Bound selection index
		if selectedIdx < 0 {
			selectedIdx = 0
		}
		if selectedIdx >= len(options) {
			selectedIdx = len(options) - 1
		}

		// Clear previous menu render
		clearLines(linesDrawn)

		// Print new menu
		var sb strings.Builder
		sb.WriteString("\n  Searching...\n\n")
		sb.WriteString("  " + ui.Bold("Nearby Devices") + "\n\n")

		for i, opt := range options {
			arrow := "   "
			if i == selectedIdx {
				arrow = " " + ui.Blue("→") + " "
			}

			if opt.isPeer {
				// Format option like: Name [status] OS [latency]
				statusStr := ui.Gray(opt.peerStatus)
				if opt.peerStatus == "Trusted" {
					statusStr = ui.Green("Trusted")
				}
				sb.WriteString(fmt.Sprintf("%s%-18s %-12s %-10s %s\n", arrow, ui.Bold(opt.label), statusStr, ui.Gray(opt.peerOS), ui.Gray(opt.peerLatency)))
				sb.WriteString(fmt.Sprintf("     %s\n", ui.Gray(opt.id)))
			} else {
				sb.WriteString(fmt.Sprintf("%s%s\n", arrow, ui.Bold(opt.label)))
			}
		}
		sb.WriteString("\n  " + ui.Gray("Use ↑/↓ to navigate, Enter to select.") + "\n")

		menuStr := sb.String()
		fmt.Print(menuStr)
		linesDrawn = strings.Count(menuStr, "\n")

		select {
		case key := <-keyCh:
			switch key {
			case "ctrlc":
				clearLines(linesDrawn)
				return fmt.Errorf("connection cancelled")
			case "up":
				selectedIdx--
			case "down":
				selectedIdx++
			case "enter":
				opt := options[selectedIdx]
				clearLines(linesDrawn)

				// Handle options
				if opt.isPeer {
					term.Restore(fd, oldState)
					return connectToPeer(opt.label, opt.id, opt.peerIP)
				} else if opt.id == "manual_id" {
					term.Restore(fd, oldState)
					fmt.Print("\n  Enter Device ID > ")
					var targetID string
					fmt.Scanln(&targetID)
					targetID = strings.TrimSpace(targetID)
					if targetID == "" {
						return fmt.Errorf("device ID cannot be empty")
					}
					// Resolve target ID from discovered peers or fail
					for _, p := range peersList {
						if p.ID == targetID {
							return connectToPeer(p.Name, p.ID, p.IP)
						}
					}
					return fmt.Errorf("device ID %s not discovered on local network", targetID)
				} else if opt.id == "manual_ip" {
					term.Restore(fd, oldState)
					fmt.Print("\n  Enter IP Address > ")
					var ipAddr string
					fmt.Scanln(&ipAddr)
					ipAddr = strings.TrimSpace(ipAddr)
					if ipAddr == "" {
						return fmt.Errorf("IP address cannot be empty")
					}
					return connectToPeer("Remote Host", "unknown-id", ipAddr)
				}
			}
		case <-ticker.C:
			refreshPeers()
		}
	}
}

func connectToPeer(name, id, ip string) error {
	fmt.Printf("\n  Connecting to %s...\n", ui.Bold(name))
	fmt.Printf("  %s\n\n", ui.Gray("Waiting for remote host to accept..."))

	// Trigger connection call to daemon
	resp, err := localClient().Post(fmt.Sprintf("https://127.0.0.1:8911/connect?ip=%s&id=%s", ip, url.QueryEscape(id)), "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to pair: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connection rejected by %s", name)
	}

	var pairResp core.PairResponse
	if err := json.NewDecoder(resp.Body).Decode(&pairResp); err != nil {
		return fmt.Errorf("malformed connection payload: %w", err)
	}

	// Success screen summary
	fmt.Fprintln(os.Stdout)
	fmt.Fprintf(os.Stdout, "  %s Connected successfully\n\n", ui.Success())
	ui.PrintDetail(os.Stdout, "Device", pairResp.DeviceName)
	ui.PrintDetail(os.Stdout, "ID", pairResp.DeviceID)
	ui.PrintDetail(os.Stdout, "Trust", "Saved")
	ui.PrintDetail(os.Stdout, "Encryption", "Enabled")
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "  Ready")
	fmt.Fprintln(os.Stdout, "    Try:")
	fmt.Fprintf(os.Stdout, "      orion run \"hostname\"\n")
	fmt.Fprintf(os.Stdout, "      orion devices\n\n")

	return nil
}
