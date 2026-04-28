package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/devsvault/devsvault/apps/cli/internal/client"
	"github.com/devsvault/devsvault/apps/cli/internal/config"
)

func newLoginCommand() *cobra.Command {
	var apiURL string
	var subject string
	var actorType string

	command := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a DevsVault API",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if subject == "" {
				return failf("--subject is required")
			}
			if actorType != "user" && actorType != "service" {
				return failf("--type must be user or service")
			}
			if _, err := config.Load(); err == nil {
				fmt.Fprint(cmd.ErrOrStderr(), "An active session already exists. Overwrite? [y/N]: ")
				answer, readErr := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
				if readErr != nil {
					return readErr
				}
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					return failf("login cancelled")
				}
			}
			apiClient := client.New(apiURL, "")
			token, expiresAt, err := apiClient.Login(subject, actorType)
			if err != nil {
				return err
			}
			if err := config.Save(config.Session{APIURL: apiURL, AccessToken: token, ExpiresAt: expiresAt}); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Logged in as %s\n", subject)
			return nil
		},
	}
	command.Flags().StringVar(&apiURL, "url", defaultAPIURL, "API URL")
	command.Flags().StringVar(&subject, "subject", "", "actor subject")
	command.Flags().StringVar(&actorType, "type", "user", "actor type: user or service")
	return command
}
