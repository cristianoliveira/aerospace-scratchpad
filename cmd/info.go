/*
Copyright © 2025 Cristian Oliveira license@cristianoliveira.dev
*/
package cmd

import (
	"fmt"

	aerospaceipc "github.com/cristianoliveira/aerospace-ipc"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/constants"
	"github.com/spf13/cobra"
)

// InfoCmd represents the info command
func InfoCmd(
	aerospace aerospaceipc.AeroSpaceClient,
) *cobra.Command {
	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "Shows relevant info about aerospace-scratchpad",
		Long: `This command provides information about the aerospace-scratchpad and aerospace.

Checks the compatibility of the installed version of Aerospace with the current version of aerospace-scratchpad.
As well as other relevant information.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			socketClient := aerospace.Connection()
			socketPath, err := socketClient.GetSocketPath()
			if err != nil {
				return fmt.Errorf("failed to get socket path: %w", err)
			}

			var validationInfo string
			if err = socketClient.CheckServerVersion(); err != nil {
				validationInfo = "Incompatible. Reason: " + err.Error()
			} else {
				validationInfo = "Compatible."
			}
			serverVersion, err := socketClient.GetServerVersion()
			if err != nil {
				return fmt.Errorf("failed to get server version: %w", err)
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
				serverVersion,
				socketPath,
				constants.DefaultScratchpadWorkspaceName,
				validationInfo,
			))

			return nil
		},
	}

	return infoCmd
}
