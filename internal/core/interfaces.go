package core

import "context"

// DeviceManager handles discovery, registration, and listing of trusted peer devices.
type DeviceManager interface {
	// AddDevice registers a new trusted remote device into the local network.
	AddDevice(dev Device) error
	// GetDevices retrieves the list of all currently paired devices.
	GetDevices() ([]Device, error)
}

// CommandRunner manages parallel command execution and real-time output streaming.
type CommandRunner interface {
	// RunCommand executes a shell command concurrently across target devices,
	// returning a read-only channel of streamed output lines.
	RunCommand(ctx context.Context, devices []Device, cmd string) <-chan ExecutionLine
}

// Engine integrates DeviceManager and CommandRunner under a unified interface.
// This interface separates the CLI layer from the actual networking implementation,
// allowing developers to easily swap out the high-fidelity mock backend for real networking.
type Engine interface {
	DeviceManager
	CommandRunner
}

// ExecutionLine represents a single line of output or error streamed from a remote device.
type ExecutionLine struct {
	DeviceName string
	Content    string
	IsStderr   bool
	Err        error
}
