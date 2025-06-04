/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"fmt"

	aerospacecli "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/spf13/cobra"
)

// InfoCmd represents the info command
func InfoCmd(
	aerospace aerospacecli.AeroSpaceClient,
) *cobra.Command {
	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "Shows relevant info about aerospace-scratchpad",
		Long: `This command provides information about the aerospace-scratchpad and aerospace.

Checks the compatibility of the installed version of Aerospace with the current version of aerospace-scratchpad.
As well as other relevant information.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			socketPath, err := aerospacecli.GetSocketPath()
			if err != nil {
				return fmt.Errorf("failed to get socket path: %w", err)
			}

			res, err := aerospace.Client().SendCommand("config", []string{"--config-path"})
			if err != nil {
				return fmt.Errorf("failed to get aerospace's config. %w", err)
			}

			var validationInfo string
			if err = aerospace.Client().CheckServerVersion(res.ServerVersion); err != nil {
				validationInfo = "Incompatible. Reason: " + err.Error()
			} else {
				validationInfo = "Compatible."
			}

			cmd.Println(fmt.Sprintf(`Aerospace Scratchpad

[Aerospace]
Version: %s
Socket: %s

[Aerospace scratchpad]
Workspace: %s

[Compatibility]
Status: %s
			`,
				res.ServerVersion,
				socketPath,
				constants.DefaultScratchpadWorkspaceName,
				validationInfo,
			))

			return nil
		},
	}

	return infoCmd
}
