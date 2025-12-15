package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/dwirx/ghex/internal/platform"
)

// Manager handles configuration loading and saving
type Manager struct {
	primaryPath string
	legacyPath  string
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	configDir := platform.GetConfigDir("ghe")
	legacyDir := platform.GetConfigDir("github-switch")

	return &Manager{
		primaryPath: filepath.Join(configDir, "config.json"),
		legacyPath:  filepath.Join(legacyDir, "config.json"),
	}
}

// GetConfigPath returns the primary configuration file path
func (m *Manager) GetConfigPath() string {
	return m.primaryPath
}

// Load reads the configuration from disk
// It tries the primary path first, then falls back to legacy path
func (m *Manager) Load() (*AppConfig, error) {
	paths := []string{m.primaryPath, m.legacyPath}

	for _, path := range paths {
		cfg, err := m.loadFromPath(path)
		if err == nil {
			// If loaded from legacy path, migrate to new location
			if path == m.legacyPath {
				_ = m.Save(cfg) // Ignore migration errors
			}
			return cfg, nil
		}

		// If file doesn't exist, try next path
		if os.IsNotExist(err) {
			continue
		}

		// For other errors, return them
		return nil, err
	}

	// No config file found, return empty config
	return NewAppConfig(), nil
}

// loadFromPath loads configuration from a specific path
func (m *Manager) loadFromPath(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Ensure accounts is not nil
	if cfg.Accounts == nil {
		cfg.Accounts = []Account{}
	}

	return &cfg, nil
}

// Save writes the configuration to disk
func (m *Manager) Save(cfg *AppConfig) error {
	// Ensure directory exists
	dir := filepath.Dir(m.primaryPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Marshal with indentation for readability
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	// Add trailing newline
	data = append(data, '\n')

	return os.WriteFile(m.primaryPath, data, 0644)
}

// Global manager instance
var defaultManager *Manager

// GetManager returns the default configuration manager
func GetManager() *Manager {
	if defaultManager == nil {
		defaultManager = NewManager()
	}
	return defaultManager
}

// Load is a convenience function to load configuration
func Load() (*AppConfig, error) {
	return GetManager().Load()
}

// Save is a convenience function to save configuration
func Save(cfg *AppConfig) error {
	return GetManager().Save(cfg)
}
