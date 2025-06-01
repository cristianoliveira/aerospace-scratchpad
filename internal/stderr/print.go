package stderr

import (
	"fmt"
	"os"

	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
)

// This package contains functions to print errors in a consistent way.
// Also handle the error codes and messages.

var ShouldExit bool = true

func SetBehavior(shouldExit bool) {
	ShouldExit = shouldExit
}

func Writef(tmpl string, a ...any) {
	logger := logger.GetDefaultLogger()
	logger.LogError(fmt.Sprintf(tmpl, a...))

	errorMessage := fmt.Errorf(tmpl, a...)
	_, err := fmt.Fprintln(os.Stderr, errorMessage)
	if err != nil {
		panic(fmt.Sprintf("Failure: unable to print error message: %v", err))
	}
	if ShouldExit {
		os.Exit(1)
	}
}

// Println prints an error message to stderr and exits the program if ShouldExit is true.
// Why not use log.Fatalf? Because we want to control the exit behavior.
func Println(tmpl string, a ...any) {
	logger := logger.GetDefaultLogger()
	logger.LogError(fmt.Sprintf(tmpl, a...))

	errorMessage := fmt.Errorf(tmpl, a...)
	_, err := fmt.Fprintln(os.Stderr, errorMessage)
	if err != nil {
		panic(fmt.Sprintf("Failure: unable to print error message: %v", err))
	}
	if ShouldExit {
		os.Exit(1)
	}
}

func Printf(tmpl string, a ...any) {
	logger := logger.GetDefaultLogger()
	logger.LogError(fmt.Sprintf(tmpl, a...))

	_, err := fmt.Fprintf(os.Stderr, tmpl, a...)
	if err != nil {
		panic(fmt.Sprintf("Failure: unable to print error message: %v", err))
	}
	if ShouldExit {
		os.Exit(1)
	}
}
