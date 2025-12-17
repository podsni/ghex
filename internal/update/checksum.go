package update

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

// CalculateChecksum computes SHA256 hash of a file
func CalculateChecksum(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// CalculateChecksumFromReader computes SHA256 hash from a reader
func CalculateChecksumFromReader(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// CalculateChecksumFromBytes computes SHA256 hash from bytes
func CalculateChecksumFromBytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// VerifyChecksum verifies file integrity using SHA256
func VerifyChecksum(filePath string, expectedChecksum string) error {
	actualChecksum, err := CalculateChecksum(filePath)
	if err != nil {
		return err
	}

	expectedChecksum = strings.ToLower(strings.TrimSpace(expectedChecksum))
	actualChecksum = strings.ToLower(actualChecksum)

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("%w: expected %s, got %s", ErrChecksumMismatch, expectedChecksum, actualChecksum)
	}

	return nil
}


// ChecksumEntry represents a single entry in a checksums file
type ChecksumEntry struct {
	Checksum string
	Filename string
}

// ParseChecksumFile parses a checksums.txt file
// Format: "checksum  filename" or "checksum filename"
func ParseChecksumFile(content string) ([]ChecksumEntry, error) {
	var entries []ChecksumEntry
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split by whitespace (could be spaces or tabs)
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		entries = append(entries, ChecksumEntry{
			Checksum: strings.ToLower(parts[0]),
			Filename: parts[len(parts)-1], // Last part is filename
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse checksum file: %w", err)
	}

	return entries, nil
}

// FindChecksum finds the checksum for a specific filename
func FindChecksum(entries []ChecksumEntry, filename string) (string, bool) {
	for _, entry := range entries {
		if entry.Filename == filename || strings.HasSuffix(entry.Filename, "/"+filename) {
			return entry.Checksum, true
		}
	}
	return "", false
}
