package config

// SshConfig holds SSH authentication configuration
type SshConfig struct {
	KeyPath   string `json:"keyPath"`
	HostAlias string `json:"hostAlias,omitempty"`
}

// TokenConfig holds token/PAT authentication configuration
type TokenConfig struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

// PlatformConfig holds git platform configuration
type PlatformConfig struct {
	Type   string `json:"type"`             // github, gitlab, bitbucket, gitea, other
	Domain string `json:"domain,omitempty"` // custom domain (e.g., gitlab.company.com)
	ApiUrl string `json:"apiUrl,omitempty"` // custom API endpoint
}

// Account represents a configured GitHub/Git account
type Account struct {
	Name        string          `json:"name"`
	GitUserName string          `json:"gitUserName,omitempty"`
	GitEmail    string          `json:"gitEmail,omitempty"`
	SSH         *SshConfig      `json:"ssh,omitempty"`
	Token       *TokenConfig    `json:"token,omitempty"`
	Platform    *PlatformConfig `json:"platform,omitempty"`
}

// HealthStatus holds the health check result for an account
type HealthStatus struct {
	AccountName string `json:"accountName"`
	SshValid    *bool  `json:"sshValid,omitempty"`
	SshError    string `json:"sshError,omitempty"`
	TokenValid  *bool  `json:"tokenValid,omitempty"`
	TokenError  string `json:"tokenError,omitempty"`
	TokenExpiry string `json:"tokenExpiry,omitempty"`
	LastChecked string `json:"lastChecked"`
}

// ActivityLogEntry represents a single activity log entry
type ActivityLogEntry struct {
	Timestamp   string `json:"timestamp"`
	Action      string `json:"action"` // switch, add, remove, edit, test
	AccountName string `json:"accountName"`
	RepoPath    string `json:"repoPath,omitempty"`
	Method      string `json:"method,omitempty"` // ssh, token
	Platform    string `json:"platform,omitempty"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
}

// AppConfig is the main application configuration
type AppConfig struct {
	Accounts        []Account          `json:"accounts"`
	ActivityLog     []ActivityLogEntry `json:"activityLog,omitempty"`
	HealthChecks    []HealthStatus     `json:"healthChecks,omitempty"`
	LastHealthCheck string             `json:"lastHealthCheck,omitempty"`
}

// NewAppConfig creates a new empty AppConfig
func NewAppConfig() *AppConfig {
	return &AppConfig{
		Accounts:     []Account{},
		ActivityLog:  []ActivityLogEntry{},
		HealthChecks: []HealthStatus{},
	}
}

// DefaultPlatform returns the default platform config (GitHub)
func DefaultPlatform() *PlatformConfig {
	return &PlatformConfig{
		Type: "github",
	}
}
