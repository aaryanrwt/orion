package core

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Capabilities advertises supported hardware/software primitives of a peer.
type Capabilities struct {
	Ollama    bool `json:"ollama"`
	CUDA      bool `json:"cuda"`
	GPU       bool `json:"gpu"`
	ROCm      bool `json:"rocm"`
	Metal     bool `json:"metal"`
	Benchmark bool `json:"benchmark"`
	Hardware  bool `json:"hardware"`
}

// OllamaModel represents an installed model tag on a machine.
type OllamaModel struct {
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	Quantization string `json:"quantization"`
	Status       string `json:"status"` // Running / Idle
}

// Device represents a paired remote machine.
type Device struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	OS              string        `json:"os"`
	IP              string        `json:"ip"`
	Latency         string        `json:"latency"`
	Status          string        `json:"status"`
	Fingerprint     string        `json:"fingerprint"` // X.509 cert fingerprint for verification
	Transport       string        `json:"transport"`   // Wi-Fi or Ethernet
	Hardware        HardwareInfo  `json:"hardware"`
	OllamaInstalled bool          `json:"ollama_installed"`
	OllamaVersion   string        `json:"ollama_version"`
	Models          []OllamaModel `json:"models"`
	Certificate     string        `json:"certificate"`
	Alias           string        `json:"alias"`
	TrustTimestamp  string        `json:"trust_timestamp"`
	Protocol        int           `json:"protocol"`
	Version         string        `json:"version"`
	Capabilities    Capabilities  `json:"capabilities"`
}

// Config holds local configuration and identity.
type Config struct {
	ConfigVersion  int      `json:"config_version"`
	DeviceID       string   `json:"device_id"`
	DeviceName     string   `json:"device_name"`
	PrivateKey     string   `json:"private_key"` // Kept for test compat
	PublicKey      string   `json:"public_key"`  // Cert fingerprint verifier
	CertificatePEM string   `json:"certificate_pem"`
	PrivateKeyPEM  string   `json:"private_key_pem"`
	Devices        []Device `json:"devices"`
	BlockedDevices []string `json:"blocked_devices"`
	LimitDevices   int      `json:"limit_devices"`
}

const configFilename = "config.json"
const configDirname = "orion"

// GetConfigPath returns the absolute path of the configuration file.
func GetConfigPath() (string, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine user config directory: %w", err)
	}
	return filepath.Join(baseDir, configDirname, configFilename), nil
}

// LoadConfig reads the configuration file from disk.
func LoadConfig() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err // Let caller handle uninitialized state
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if cfg.CertificatePEM == "" && cfg.DeviceID != "" {
		// Automatically generate certificates for existing configs
		privPEM, certPEM, fingerprint, err := GenerateIdentityCert(cfg.DeviceID)
		if err == nil {
			cfg.CertificatePEM = certPEM
			cfg.PrivateKeyPEM = privPEM
			cfg.PublicKey = fingerprint
			cfg.PrivateKey = privPEM
			SaveConfig(&cfg)
		}
	}

	return &cfg, nil
}

// SaveConfig writes the configuration file to disk.
func SaveConfig(cfg *Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GenerateRandomHex generates a random hex string of specified byte length.
func GenerateRandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// InitializeLocalDevice creates and saves a brand new configuration.
func InitializeLocalDevice() (*Config, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-host"
	}

	part1, err := GenerateRandomHex(2)
	if err != nil {
		return nil, err
	}
	part2, err := GenerateRandomHex(2)
	if err != nil {
		return nil, err
	}
	deviceID := fmt.Sprintf("ORN-%s-%s", strings.ToUpper(part1), strings.ToUpper(part2))

	privPEM, certPEM, fingerprint, err := GenerateIdentityCert(deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate identity cert: %w", err)
	}

	cfg := &Config{
		ConfigVersion:  1,
		DeviceID:       deviceID,
		DeviceName:     hostname,
		PrivateKey:     privPEM,
		PublicKey:      fingerprint,
		CertificatePEM: certPEM,
		PrivateKeyPEM:  privPEM,
		Devices:        []Device{},
		BlockedDevices: []string{},
		LimitDevices:   5,
	}

	if err := SaveConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// RenameLocalDevice updates the editable local name in config.
func RenameLocalDevice(newName string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	cfg.DeviceName = newName
	return SaveConfig(cfg)
}

func stringsToAlphaNumeric(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			sb.WriteRune(r)
		}
	}
	return strings.ToLower(sb.String())
}
