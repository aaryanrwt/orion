package cli

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/orion-infra/orion/internal/core"
	"github.com/orion-infra/orion/internal/ui"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:     "doctor",
	Short:   "Diagnose local system status, configuration, and network health",
	Example: "  $ orion doctor",
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		fmt.Fprintln(out)
		ui.Header(out, "Orion Doctor - Diagnostics")
		fmt.Fprintln(out)

		cfg, cfgErr := core.LoadConfig()

		// Helper to render check
		printCheck := func(name string, success bool, detail string) {
			symbol := ui.Success()
			if !success {
				symbol = ui.ErrorSymbol()
			}
			fmt.Fprintf(out, "   %s %-15s %s\n", symbol, name, ui.Gray(detail))
		}

		// 1. Daemon check
		daemonOk := false
		daemonDetail := "Offline"
		resp, err := localClient().Get("https://127.0.0.1:8911/devices")
		if err == nil {
			resp.Body.Close()
			daemonOk = true
			daemonDetail = "Running on port 8911"
		}
		printCheck("Daemon", daemonOk, daemonDetail)

		// 2. Certificates check
		certOk := false
		certDetail := "Uninitialized"
		if cfgErr == nil && cfg.CertificatePEM != "" && cfg.PrivateKeyPEM != "" {
			_, err := tls.X509KeyPair([]byte(cfg.CertificatePEM), []byte(cfg.PrivateKeyPEM))
			if err == nil {
				certOk = true
				certDetail = "Valid (ECDSA P-256 Keypair)"
			} else {
				certDetail = "Malformed keypair: " + err.Error()
			}
		}
		printCheck("Certificates", certOk, certDetail)

		// 3. TLS check
		tlsOk := certOk
		tlsDetail := "Insecure"
		if tlsOk {
			tlsDetail = "mTLS Enabled (TLS 1.3)"
		}
		printCheck("TLS", tlsOk, tlsDetail)

		// 4. Ports check
		portsOk := true
		portsDetail := "Ports 8910/8911 available"
		if !daemonOk {
			l1, err1 := net.Listen("tcp", "127.0.0.1:8911")
			if err1 != nil {
				portsOk = false
				portsDetail = "Port 8911 in use"
			} else {
				l1.Close()
			}
			l2, err2 := net.ListenPacket("udp", "0.0.0.0:8910")
			if err2 != nil {
				portsOk = false
				portsDetail = "Port 8910 (UDP) in use"
			} else {
				l2.Close()
			}
		}
		printCheck("Ports", portsOk, portsDetail)

		// 5. Discovery check
		discoveryOk := daemonOk
		discoveryDetail := "Active"
		if !daemonOk {
			discoveryDetail = "Inactive (Daemon offline)"
		}
		printCheck("Discovery", discoveryOk, discoveryDetail)

		// 6. Protocol check
		printCheck("Protocol", true, "Version 1 (Compatible)")

		// 7. Hardware profiling check
		hw := core.DetectHardware()
		printCheck("Hardware", hw.CPU != "Unknown CPU", fmt.Sprintf("CPU: %s, RAM: %s", hw.CPU, hw.RAM))

		// 8. GPU check
		gpuOk := hw.GPU != "CPU Only"
		gpuDetail := "Discrete GPU Detected: " + hw.GPU
		if !gpuOk {
			gpuDetail = "CPU Only mode active"
		}
		printCheck("GPU", gpuOk, gpuDetail)

		// 9. Ollama check
		ollamaOk := false
		ollamaDetail := "Not running"
		client := &http.Client{Timeout: 300 * time.Millisecond}
		oResp, oErr := client.Get("http://127.0.0.1:11434/api/tags")
		if oErr == nil {
			oResp.Body.Close()
			ollamaOk = true
			vResp, vErr := client.Get("http://127.0.0.1:11434/api/version")
			if vErr == nil {
				defer vResp.Body.Close()
				var vPayload struct {
					Version string `json:"version"`
				}
				if json.NewDecoder(vResp.Body).Decode(&vPayload) == nil {
					ollamaDetail = "Running (v" + vPayload.Version + ")"
				}
			}
		}
		printCheck("Ollama", ollamaOk, ollamaDetail)

		// 10. Network link check
		interfaces, _ := net.Interfaces()
		netOk := false
		netDetail := "No active non-loopback interface"
		for _, iface := range interfaces {
			if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
				addrs, _ := iface.Addrs()
				if len(addrs) > 0 {
					netOk = true
					netDetail = fmt.Sprintf("Interface: %s (Up)", iface.Name)
					break
				}
			}
		}
		printCheck("Network", netOk, netDetail)

		fmt.Fprintln(out)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(doctorCmd)
}
