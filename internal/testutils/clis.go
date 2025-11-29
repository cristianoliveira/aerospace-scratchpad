package testutils

// This module contains test utilities for CLI commands.
// - Shell output
// - Cobra command execution
// - Standard input/output capturing

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc/pkg/aerospace"
)

func CmdExecute(cmd *cobra.Command, args ...string) (string, error) {
	cmd.SetArgs(args)
	stdOut, err := CaptureStdOut(func() error {
		return cmd.Execute()
	})

	if err != nil {
		return "", err
	}

	return stdOut, nil
}

//nolint:reassign // CaptureStdOut temporarily redirects standard streams for testing
func CaptureStdOut(f func() error) (string, error) {
	var buf bytes.Buffer
	// Save original stdout
	old := os.Stdout
	oldErr := os.Stderr
	// Redirect stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	errFile, _ := os.CreateTemp("", "aerospace-scratchpad-stdout")
	os.Stderr = errFile // Redirect stderr to the same pipe

	// Run the function that prints to stdout
	err := f()
	if err != nil {
		return "", err
	}

	err = errFile.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close error file: %w", err)
	}
	// Restore stderr to original
	os.Stderr = oldErr

	// read the error file
	errFileContent, err := os.ReadFile(errFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read error file: %w", err)
	}
	if len(errFileContent) > 0 {
		return "", fmt.Errorf("%s", errFileContent)
	}

	// Close writer and restore stdout
	err = w.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}
	os.Stdout = old

	// Read output
	_, err = io.Copy(&buf, r)
	if err != nil {
		return "", fmt.Errorf("failed to read output: %w", err)
	}
	return buf.String(), nil
}

type MockEmptyAerspaceMarkWindows struct{}

func (d *MockEmptyAerspaceMarkWindows) Client() *aerospacecli.AeroSpaceWM {
	return &aerospacecli.AeroSpaceWM{}
}
