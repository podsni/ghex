package update

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected *Version
		wantErr  bool
	}{
		{"1.0.0", &Version{Major: 1, Minor: 0, Patch: 0}, false},
		{"v1.0.0", &Version{Major: 1, Minor: 0, Patch: 0}, false},
		{"2.3.4", &Version{Major: 2, Minor: 3, Patch: 4}, false},
		{"v2.3.4", &Version{Major: 2, Minor: 3, Patch: 4}, false},
		{"1.0.0-beta.1", &Version{Major: 1, Minor: 0, Patch: 0, Pre: "beta.1"}, false},
		{"v1.0.0-rc.1", &Version{Major: 1, Minor: 0, Patch: 0, Pre: "rc.1"}, false},
		{"invalid", nil, true},
		{"1.0", nil, true},
		{"", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Major != tt.expected.Major || got.Minor != tt.expected.Minor ||
					got.Patch != tt.expected.Patch || got.Pre != tt.expected.Pre {
					t.Errorf("ParseVersion(%q) = %+v, want %+v", tt.input, got, tt.expected)
				}
			}
		})
	}
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"1.1.0", "1.0.0", 1},
		{"2.0.0", "1.9.9", 1},
		{"1.0.0", "1.0.0-beta.1", 1},  // Release > pre-release
		{"1.0.0-beta.1", "1.0.0", -1}, // Pre-release < release
		{"1.0.0-alpha", "1.0.0-beta", -1},
	}

	for _, tt := range tests {
		t.Run(tt.v1+" vs "+tt.v2, func(t *testing.T) {
			v1, _ := ParseVersion(tt.v1)
			v2, _ := ParseVersion(tt.v2)
			got := v1.Compare(v2)
			if got != tt.expected {
				t.Errorf("Compare(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.expected)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		version  *Version
		expected string
	}{
		{&Version{Major: 1, Minor: 0, Patch: 0}, "1.0.0"},
		{&Version{Major: 2, Minor: 3, Patch: 4}, "2.3.4"},
		{&Version{Major: 1, Minor: 0, Patch: 0, Pre: "beta.1"}, "1.0.0-beta.1"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.version.String()
			if got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestVersionIsNewerThan(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected bool
	}{
		{"2.0.0", "1.0.0", true},
		{"1.0.0", "2.0.0", false},
		{"1.0.0", "1.0.0", false},
		{"1.0.1", "1.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.v1+" > "+tt.v2, func(t *testing.T) {
			v1, _ := ParseVersion(tt.v1)
			v2, _ := ParseVersion(tt.v2)
			got := v1.IsNewerThan(v2)
			if got != tt.expected {
				t.Errorf("IsNewerThan(%q, %q) = %v, want %v", tt.v1, tt.v2, got, tt.expected)
			}
		})
	}
}
