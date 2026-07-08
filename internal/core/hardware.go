package core

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// HardwareInfo stores physical resource specs of a node.
type HardwareInfo struct {
	CPU           string `json:"cpu"`
	RAM           string `json:"ram"`
	GPU           string `json:"gpu"`
	VRAM          string `json:"vram"`
	CUDA          string `json:"cuda"`
	OS            string `json:"os"`
	Arch          string `json:"arch"`
	DriverVersion string `json:"driver_version"`
	CUDAVersion   string `json:"cuda_version"`
}

// DetectHardware queries OS-specific commands to profile the machine.
func DetectHardware() HardwareInfo {
	info := HardwareInfo{
		CPU:           "Unknown CPU",
		RAM:           "Unknown RAM",
		GPU:           "CPU Only",
		VRAM:          "-",
		CUDA:          "No",
		OS:            runtime.GOOS,
		Arch:          runtime.GOARCH,
		DriverVersion: "-",
		CUDAVersion:   "-",
	}

	// 1. Profiling CPU & RAM
	switch runtime.GOOS {
	case "windows":
		// RAM
		cmd := exec.Command("powershell", "-Command", "[math]::round((Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory / 1GB)")
		var out bytes.Buffer
		cmd.Stdout = &out
		if cmd.Run() == nil {
			info.RAM = strings.TrimSpace(out.String()) + " GB"
		}
		// CPU
		out.Reset()
		cmd = exec.Command("powershell", "-Command", "(Get-CimInstance Win32_Processor).Name")
		cmd.Stdout = &out
		if cmd.Run() == nil {
			info.CPU = strings.TrimSpace(out.String())
		}

	case "darwin":
		// RAM
		cmd := exec.Command("sysctl", "-n", "hw.memsize")
		var out bytes.Buffer
		cmd.Stdout = &out
		if cmd.Run() == nil {
			mem, _ := strconv.ParseUint(strings.TrimSpace(out.String()), 10, 64)
			info.RAM = fmt.Sprintf("%d GB", mem/1024/1024/1024)
		}
		// CPU (which checks if M-series Apple Silicon)
		out.Reset()
		cmd = exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
		cmd.Stdout = &out
		if cmd.Run() == nil {
			info.CPU = strings.TrimSpace(out.String())
			if strings.Contains(info.CPU, "Apple") {
				info.GPU = "Apple Unified Graphics"
				info.VRAM = info.RAM // Unified memory
			}
		}

	case "linux":
		// RAM
		cmd := exec.Command("sh", "-c", "free -g | awk '/^Mem:/{print $2}'")
		var out bytes.Buffer
		cmd.Stdout = &out
		if cmd.Run() == nil {
			info.RAM = strings.TrimSpace(out.String()) + " GB"
		}
		// CPU
		out.Reset()
		cmd = exec.Command("sh", "-c", "lscpu | grep 'Model name' | cut -d: -f2")
		cmd.Stdout = &out
		if cmd.Run() == nil {
			info.CPU = strings.TrimSpace(out.String())
		}
	}

	// Clean strings
	info.CPU = strings.TrimSpace(info.CPU)
	info.RAM = strings.TrimSpace(info.RAM)

	// 2. Profiling NVIDIA GPU (works on Windows & Linux if nvidia-smi is installed)
	cmd := exec.Command("nvidia-smi", "--query-gpu=gpu_name,memory.total", "--format=csv,noheader,nounits")
	var out bytes.Buffer
	cmd.Stdout = &out
	if cmd.Run() == nil {
		parts := strings.Split(strings.TrimSpace(out.String()), ",")
		if len(parts) >= 2 {
			info.GPU = strings.TrimSpace(parts[0])
			vramMB, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err == nil {
				info.VRAM = fmt.Sprintf("%.1f GB", vramMB/1024)
			}
			info.CUDA = "Yes"

			// Query Driver Version
			cmdDrv := exec.Command("nvidia-smi", "--query-gpu=driver_version", "--format=csv,noheader")
			var outDrv bytes.Buffer
			cmdDrv.Stdout = &outDrv
			if cmdDrv.Run() == nil {
				info.DriverVersion = strings.TrimSpace(outDrv.String())
			}

			// Query full nvidia-smi to parse CUDA version
			cmdFull := exec.Command("nvidia-smi")
			var outFull bytes.Buffer
			cmdFull.Stdout = &outFull
			if cmdFull.Run() == nil {
				fullStr := outFull.String()
				idx := strings.Index(fullStr, "CUDA Version:")
				if idx != -1 {
					sub := fullStr[idx+len("CUDA Version:"):]
					pFields := strings.Fields(sub)
					if len(pFields) > 0 {
						info.CUDAVersion = pFields[0]
					}
				}
			}
		}
	}

	return info
}
