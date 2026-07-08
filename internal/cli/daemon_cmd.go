package cli

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/orion-infra/orion/internal/core"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:    "daemon",
	Short:  "Start the Orion background mesh service",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := core.LoadConfig()
		if err != nil {
			return fmt.Errorf("daemon initialization failed (uninitialized): %w", err)
		}

		daemon, err := core.NewDaemon(cfg)
		if err != nil {
			return err
		}

		return daemon.Start(false)
	},
}

func init() {
	RootCmd.AddCommand(daemonCmd)
}

// localClient returns an HTTPS client configured for local loopback communication.
func localClient() *http.Client {
	var certs []tls.Certificate
	cfg, err := core.LoadConfig()
	if err == nil && cfg.CertificatePEM != "" {
		cert, err := tls.X509KeyPair([]byte(cfg.CertificatePEM), []byte(cfg.PrivateKeyPEM))
		if err == nil {
			certs = append(certs, cert)
		}
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       certs,
				InsecureSkipVerify: true,
			},
		},
		Timeout: 5 * time.Second,
	}
}

// ensureDaemonRunning validates daemon status and spawns a background service if inactive.
func ensureDaemonRunning() error {
	// Check if config exists. If not, do not attempt to start daemon.
	cfg, err := core.LoadConfig()
	if err != nil {
		return fmt.Errorf("orion is not initialized: %w", err)
	}

	var certs []tls.Certificate
	if cfg.CertificatePEM != "" {
		cert, err := tls.X509KeyPair([]byte(cfg.CertificatePEM), []byte(cfg.PrivateKeyPEM))
		if err == nil {
			certs = append(certs, cert)
		}
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       certs,
				InsecureSkipVerify: true,
			},
		},
		Timeout: 200 * time.Millisecond,
	}

	// Try pinging local daemon status
	_, err = client.Get("https://127.0.0.1:8911/devices")
	if err == nil {
		return nil // Daemon already active
	}

	// Spawn invisible background daemon
	exePath, err := os.Executable()
	if err != nil {
		exePath = os.Args[0]
	}

	cmd := exec.Command(exePath, "daemon")
	hideWindow(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start background daemon: %w", err)
	}

	// Wait up to 1.5 seconds for daemon to begin listening
	for i := 0; i < 15; i++ {
		time.Sleep(100 * time.Millisecond)
		_, err = client.Get("https://127.0.0.1:8911/devices")
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("background daemon listener failed to start on port 8911")
}
