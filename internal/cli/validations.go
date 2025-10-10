package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// Validations package contains functions to help
// validate the command line arguments and options.

// ValidateAllNonEmpty checks if all arguments are non-empty.
func ValidateAllNonEmpty(_ *cobra.Command, args []string) error {
	for i, arg := range args {
		if strings.TrimSpace(arg) == "" {
			return fmt.Errorf(
				"argument at position %d is empty or whitespace",
				i,
			)
		}
	}

	return nil
}
