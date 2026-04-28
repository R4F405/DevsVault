package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/devsvault/devsvault/apps/cli/internal/config"
)

func newLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove the local DevsVault session",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := config.Clear(); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Logged out")
			return nil
		},
	}
}
