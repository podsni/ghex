package shell

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Run executes a command and returns the output
// It returns an error if the command fails
func Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Include stderr in error message
		errMsg := stderr.String()
		if errMsg != "" {
			return stdout.String(), fmt.Errorf("%w: %s", err, strings.TrimSpace(errMsg))
		}
		return stdout.String(), err
	}

	return strings.TrimSpace(stdout.String()), nil
}

// RunInDir executes a command in a specific directory
func RunInDir(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := stderr.String()
		if errMsg != "" {
			return stdout.String(), fmt.Errorf("%w: %s", err, strings.TrimSpace(errMsg))
		}
		return stdout.String(), err
	}

	return strings.TrimSpace(stdout.String()), nil
}

// Exec executes a command and returns combined stdout and stderr
// It doesn't return an error for non-zero exit codes (useful for commands like ssh -T)
func Exec(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()
	return output.String(), err
}

// ExecInDir executes a command in a specific directory
func ExecInDir(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()
	return output.String(), err
}

// RunInteractive runs a command with stdin/stdout/stderr connected to the terminal
func RunInteractive(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// RunInteractiveInDir runs an interactive command in a specific directory
func RunInteractiveInDir(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// CommandExists checks if a command exists in PATH
func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// GetExitCode returns the exit code from an error
func GetExitCode(err error) int {
	if err == nil {
		return 0
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}

	return -1
}
