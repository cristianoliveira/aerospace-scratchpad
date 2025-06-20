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

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/spf13/cobra"
)

func CmdExecute(cmd *cobra.Command, args ...string) (string, error) {
	cmd.SetArgs(args)
	stdOut, err := CaptureStdOut(func() error {
		return cmd.Execute()
	})

	if err != nil {
		return "", err
	}

	return string(stdOut), nil
}

func CaptureStdOut(f func() error) (string, error) {
	var buf bytes.Buffer
	// Save original stdout
	old := os.Stdout
	// Redirect stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w // Redirect stderr to the same pipe

	// Run the function that prints to stdout
	err := f()
	if err != nil {
		return "", err
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

func (d *MockEmptyAerspaceMarkWindows) GetWindowByID(windowID string) (*aerospacecli.Window, error) {
	fmt.Println("Mocked GetWindowByID called with windowID:", windowID)
	return &aerospacecli.Window{}, nil
}
