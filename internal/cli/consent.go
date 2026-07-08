package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/orion-infra/orion/internal/ui"
	"golang.org/x/term"
)

// HandleConsentPrompt displays an interactive prompt asking the user to accept/reject a pairing request.
func HandleConsentPrompt(deviceName, deviceID, fingerprint string) error {
	fmt.Fprintln(os.Stdout)
	ui.Header(os.Stdout, "Incoming Connection Request")
	fmt.Fprintln(os.Stdout)

	ui.PrintDetail(os.Stdout, "Device", deviceName)
	ui.PrintDetail(os.Stdout, "ID", deviceID)
	ui.PrintDetail(os.Stdout, "Fingerprint", fingerprint)
	fmt.Fprintln(os.Stdout)

	fmt.Fprint(os.Stdout, "  Accept? [Y] Accept  [N] Reject > ")

	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer term.Restore(fd, oldState)

	var buf [1]byte
	var action string
	for {
		_, err := os.Stdin.Read(buf[:])
		if err != nil {
			return err
		}
		char := strings.ToLower(string(buf[:]))
		if char == "y" {
			action = "accept"
			break
		}
		if char == "n" || buf[0] == 3 { // 'n' or Ctrl+C
			action = "reject"
			break
		}
	}

	term.Restore(fd, oldState)
	fmt.Fprintln(os.Stdout)

	// Send decision to daemon
	respondResp, err := localClient().Post(fmt.Sprintf("https://127.0.0.1:8911/respond?action=%s", action), "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to submit response: %w", err)
	}
	defer respondResp.Body.Close()

	fmt.Fprintln(os.Stdout)
	if action == "accept" {
		fmt.Fprintf(os.Stdout, "  %s Trust established with %s\n\n", ui.Success(), deviceName)
	} else {
		fmt.Fprintf(os.Stdout, "  %s Connection rejected\n\n", ui.ErrorSymbol())
		os.Exit(0) // Exit on reject to prevent proceeding with command
	}

	return nil
}
