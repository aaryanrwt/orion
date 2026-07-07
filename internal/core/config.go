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

// Device represents a paired remote machine.
type Device struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	OS      string `json:"os"`
	IP      string `json:"ip"`
	Latency string `json:"latency"`
	Status  string `json:"status"`
}

// Config holds local configuration and identity.
type Config struct {
	DeviceID   string   `json:"device_id"`
	DeviceName string   `json:"device_name"`
	PrivateKey string   `json:"private_key"`
	PublicKey  string   `json:"public_key"`
	Devices    []Device `json:"devices"`
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
	// Sanitize hostname to fit identifier scheme (lowercase, alpha-numeric)
	hostname = stringsToAlphaNumeric(hostname)

	suffix, err := GenerateRandomHex(4)
	if err != nil {
		return nil, fmt.Errorf("failed to generate device id suffix: %w", err)
	}
	deviceID := fmt.Sprintf("%s-%s", hostname, suffix)

	privKey, err := GenerateRandomHex(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}
	pubKey, err := GenerateRandomHex(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate public key: %w", err)
	}

	cfg := &Config{
		DeviceID:   deviceID,
		DeviceName: hostname,
		PrivateKey: privKey,
		PublicKey:  pubKey,
		Devices:    []Device{},
	}

	if err := SaveConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// stringsToAlphaNumeric sanitizes input strings to make them safe identifiers.
func stringsToAlphaNumeric(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			sb.WriteRune(r)
		}
	}
	return strings.ToLower(sb.String())
}
