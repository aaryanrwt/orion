package core

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// mockEngine implements the Engine interface for CLI simulation and real local process execution.
type mockEngine struct {
	cfg *Config
}

// NewMockEngine instantiates a simulated execution and device pairing engine.
func NewMockEngine(cfg *Config) Engine {
	return &mockEngine{cfg: cfg}
}

// AddDevice saves a new trusted peer device to the configuration file.
func (m *mockEngine) AddDevice(dev Device) error {
	m.cfg.Devices = append(m.cfg.Devices, dev)
	return SaveConfig(m.cfg)
}

// GetDevices returns the list of paired peer devices from the configuration.
func (m *mockEngine) GetDevices() ([]Device, error) {
	return m.cfg.Devices, nil
}

// RunCommand runs the command for real locally, and simulated for remote nodes.
func (m *mockEngine) RunCommand(ctx context.Context, devices []Device, cmd string) <-chan ExecutionLine {
	ch := make(chan ExecutionLine)
	go m.simulateExecution(ctx, devices, cmd, ch)
	return ch
}

func (m *mockEngine) simulateExecution(ctx context.Context, devices []Device, cmd string, ch chan<- ExecutionLine) {
	defer close(ch)
	var wg sync.WaitGroup

	for _, d := range devices {
		wg.Add(1)
		go func(dev Device) {
			defer wg.Done()

			// Check if it is the local machine (initiator)
			if dev.ID == m.cfg.DeviceID {
				args := strings.Fields(cmd)
				if len(args) == 0 {
					ch <- ExecutionLine{DeviceName: dev.Name, Err: fmt.Errorf("empty command")}
					return
				}

				// Run local command for real using os/exec!
				localCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
				defer cancel()

				c := exec.CommandContext(localCtx, args[0], args[1:]...)
				stdout, err := c.StdoutPipe()
				if err != nil {
					ch <- ExecutionLine{DeviceName: dev.Name, Err: err}
					return
				}
				stderr, err := c.StderrPipe()
				if err != nil {
					ch <- ExecutionLine{DeviceName: dev.Name, Err: err}
					return
				}

				if err := c.Start(); err != nil {
					ch <- ExecutionLine{DeviceName: dev.Name, Err: fmt.Errorf("failed to start: %w", err)}
					return
				}

				var wgRead sync.WaitGroup
				wgRead.Add(2)

				readStream := func(r io.Reader, isStderr bool) {
					defer wgRead.Done()
					scanner := bufio.NewScanner(r)
					for scanner.Scan() {
						ch <- ExecutionLine{
							DeviceName: dev.Name,
							Content:    scanner.Text(),
							IsStderr:   isStderr,
						}
					}
				}

				go readStream(stdout, false)
				go readStream(stderr, true)

				wgRead.Wait()
				if err := c.Wait(); err != nil {
					ch <- ExecutionLine{DeviceName: dev.Name, Err: err}
				}
				return
			}

			// Remote devices: Simulation Warning
			select {
			case <-ctx.Done():
				ch <- ExecutionLine{DeviceName: dev.Name, Err: ctx.Err()}
				return
			case <-time.After(200 * time.Millisecond):
			}

			if dev.Status == "offline" {
				ch <- ExecutionLine{
					DeviceName: dev.Name,
					Err:        fmt.Errorf("connection timeout: device %s is offline", dev.Name),
				}
				return
			}

			warnings := []string{
				"Execution engine unavailable.",
				"Reason: Version 0.1 currently uses a local simulation. This feature will become available in v0.2.",
				"Learn more: https://orion.sh/docs/simulation",
			}

			for _, w := range warnings {
				ch <- ExecutionLine{
					DeviceName: dev.Name,
					Content:    w,
					IsStderr:   true,
				}
			}
		}(d)
	}

	wg.Wait()
}
