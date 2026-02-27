package account

import (
	"os"
	"time"

	"github.com/dwirx/ghex/internal/config"
	"github.com/dwirx/ghex/internal/platform"
)

// Health indicator symbols
const (
	HealthValid   = "✓"
	HealthInvalid = "✗"
	HealthUnknown = "?"
)

// StaleThreshold is the duration after which health data is considered stale
const StaleThreshold = 24 * time.Hour

// HealthIndicators contains health check indicators for an account
type HealthIndicators struct {
	SSHKeyExists bool
	SSHKeyValid  *bool // nil = unknown
	TokenValid   *bool // nil = unknown
	LastChecked  time.Time
	IsStale      bool // true if > 24 hours old
}

// HealthState represents the state of a health check
type HealthState int

const (
	HealthStateUnknown HealthState = iota
	HealthStateValid
	HealthStateInvalid
)

// GetHealthIndicator returns the correct symbol for a health state
func GetHealthIndicator(state HealthState) string {
	switch state {
	case HealthStateValid:
		return HealthValid
	case HealthStateInvalid:
		return HealthInvalid
	default:
		return HealthUnknown
	}
}

// GetHealthIndicatorFromBool returns indicator from bool pointer
func GetHealthIndicatorFromBool(valid *bool) string {
	if valid == nil {
		return HealthUnknown
	}
	if *valid {
		return HealthValid
	}
	return HealthInvalid
}

// CheckSSHKeyHealth verifies SSH key file exists and is valid
func CheckSSHKeyHealth(keyPath string) HealthIndicators {
	indicators := HealthIndicators{
		LastChecked: time.Now(),
		IsStale:     false,
	}

	// Expand path (handle ~)
	expandedPath := platform.ExpandPath(keyPath)

	// Check if file exists
	info, err := os.Stat(expandedPath)
	if err != nil {
		indicators.SSHKeyExists = false
		invalid := false
		indicators.SSHKeyValid = &invalid
		return indicators
	}

	indicators.SSHKeyExists = true

	// Check if it's a regular file (not directory)
	if info.IsDir() {
		invalid := false
		indicators.SSHKeyValid = &invalid
		return indicators
	}

	// Check file permissions (should not be world-readable)
	// On Windows, permission bits are meaningless (always return 0666/0777)
	if !platform.IsWindows() {
		mode := info.Mode()
		if mode&0077 != 0 {
			// File has group or world permissions - potentially insecure but still valid
			valid := true
			indicators.SSHKeyValid = &valid
			return indicators
		}
	}

	valid := true
	indicators.SSHKeyValid = &valid
	return indicators
}

// CheckTokenHealth placeholder for token validation
// Note: Actual token validation requires API call to the platform
func CheckTokenHealth(token *config.TokenConfig, platformType string) HealthIndicators {
	indicators := HealthIndicators{
		LastChecked: time.Now(),
		IsStale:     false,
	}

	if token == nil || token.Token == "" {
		invalid := false
		indicators.TokenValid = &invalid
		return indicators
	}

	// Token exists but we can't validate without API call
	// Mark as unknown
	indicators.TokenValid = nil
	return indicators
}

// IsStaleCheck checks if health data is older than 24 hours
func IsStaleCheck(lastChecked time.Time) bool {
	if lastChecked.IsZero() {
		return true
	}
	return time.Since(lastChecked) > StaleThreshold
}

// GetAccountHealth returns health indicators for an account
func GetAccountHealth(account config.Account, healthStatus *config.HealthStatus) HealthIndicators {
	indicators := HealthIndicators{
		IsStale: true,
	}

	// Check SSH key health
	if account.SSH != nil && account.SSH.KeyPath != "" {
		sshHealth := CheckSSHKeyHealth(account.SSH.KeyPath)
		indicators.SSHKeyExists = sshHealth.SSHKeyExists
		indicators.SSHKeyValid = sshHealth.SSHKeyValid
	}

	// Use cached health status if available
	if healthStatus != nil {
		indicators.TokenValid = healthStatus.TokenValid

		// Parse last checked time
		if healthStatus.LastChecked != "" {
			if t, err := time.Parse(time.RFC3339, healthStatus.LastChecked); err == nil {
				indicators.LastChecked = t
				indicators.IsStale = IsStaleCheck(t)
			}
		}
	}

	return indicators
}

// FormatHealthDisplay returns formatted health display string
func FormatHealthDisplay(indicators HealthIndicators) string {
	result := ""

	// SSH health
	if indicators.SSHKeyValid != nil {
		result += "SSH:" + GetHealthIndicatorFromBool(indicators.SSHKeyValid)
	}

	// Token health
	if indicators.TokenValid != nil {
		if result != "" {
			result += " "
		}
		result += "Token:" + GetHealthIndicatorFromBool(indicators.TokenValid)
	} else if result == "" {
		result = HealthUnknown
	}

	// Stale indicator
	if indicators.IsStale && !indicators.LastChecked.IsZero() {
		result += " (stale)"
	}

	return result
}
