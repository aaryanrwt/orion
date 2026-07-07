package cli

import (
	"bytes"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/orion-infra/orion/internal/ui"
)

func init() {
	ui.NoColor = true
}

func setupTestConfig(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "orion-cli-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", tempDir)
	} else {
		t.Setenv("XDG_CONFIG_HOME", tempDir)
		t.Setenv("HOME", tempDir)
	}
	return tempDir
}

func executeCmd(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs(args)

	err := RootCmd.Execute()
	return buf.String(), err
}

func TestVersionCommand(t *testing.T) {
	output, err := executeCmd("version")
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	if !strings.Contains(output, "v0.1.0-beta") {
		t.Errorf("unexpected version output: %s", output)
	}
}

func TestFirstTimeDeveloperJourney(t *testing.T) {
	// 1. Establish fresh config folder sandbox
	setupTestConfig(t)

	// 2. Run devices before init - should fail
	_, err := executeCmd("devices")
	if err == nil {
		t.Error("expected devices to fail before initialization, but got nil error")
	}

	// 3. Initialize Orion local node
	initOut, err := executeCmd("init")
	if err != nil {
		t.Fatalf("init command failed: %v", err)
	}
	if !strings.Contains(initOut, "Orion initialized successfully") {
		t.Errorf("unexpected init output: %s", initOut)
	}

	// 4. Run status
	statusOut, err := executeCmd("status")
	if err != nil {
		t.Fatalf("status command failed: %v", err)
	}
	if !strings.Contains(statusOut, "Connected") || !strings.Contains(statusOut, "1 device (local)") {
		t.Errorf("unexpected status output: %s", statusOut)
	}

	// 5. Join a mock device
	joinOut, err := executeCmd("join", "sirius-99a3")
	if err != nil {
		t.Fatalf("join command failed: %v", err)
	}
	if !strings.Contains(joinOut, "Secure connection established") || !strings.Contains(joinOut, "Added device \"sirius\"") {
		t.Errorf("unexpected join output: %s", joinOut)
	}

	// 6. Verify devices table lists both initiator and remote
	devicesOut, err := executeCmd("devices")
	if err != nil {
		t.Fatalf("devices command failed: %v", err)
	}
	if !strings.Contains(devicesOut, "DEVICE ID") || !strings.Contains(devicesOut, "sirius") {
		t.Errorf("unexpected devices output: %s", devicesOut)
	}

	// 7. Execute command
	runOut, err := executeCmd("run", "go version")
	if err != nil {
		t.Fatalf("run command failed: %v", err)
	}
	if !strings.Contains(runOut, "success") || !strings.Contains(runOut, "simulated") {
		t.Errorf("unexpected run output: %s", runOut)
	}

	// 8. Run doctor diagnostics
	doctorOut, err := executeCmd("doctor")
	if err != nil {
		t.Fatalf("doctor command failed: %v", err)
	}
	if !strings.Contains(doctorOut, "Orion Doctor") || !strings.Contains(doctorOut, "CLI") {
		t.Errorf("unexpected doctor output: %s", doctorOut)
	}
}
