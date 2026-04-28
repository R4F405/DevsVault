package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/devsvault/devsvault/apps/cli/internal/client"
	"github.com/devsvault/devsvault/apps/cli/internal/config"
)

func newSecretsCommand() *cobra.Command {
	command := &cobra.Command{Use: "secrets", Short: "Manage secrets"}
	command.AddCommand(newSecretsListCommand(), newSecretsGetCommand(), newSecretsSetCommand(), newSecretsRotateCommand(), newSecretsRevokeCommand())
	return command
}

func newSecretsListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List secret metadata",
		RunE: func(cmd *cobra.Command, _ []string) error {
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			items, err := apiClient.ListSecrets()
			if err != nil {
				return err
			}
			if len(items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No secrets found")
				return nil
			}
			writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(writer, "PATH\tVERSION\tUPDATED")
			for _, item := range items {
				fmt.Fprintf(writer, "%s\t%d\t%s\n", item.LogicalPath, item.ActiveVersion, formatTime(item.UpdatedAt))
			}
			return writer.Flush()
		},
	}
}

func newSecretsGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <path>",
		Short: "Print a secret value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			value, err := apiClient.GetSecret(args[0])
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), value)
			return err
		},
	}
}

func newSecretsSetCommand() *cobra.Command {
	var workspace string
	var project string
	var env string
	command := &cobra.Command{
		Use:   "set <path> <value>",
		Short: "Create or rotate a secret",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			value := args[1]
			parts, err := parseSecretPath(path, workspace, project, env)
			if err != nil {
				return err
			}
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			items, err := apiClient.ListSecrets()
			if err != nil {
				return err
			}
			for _, item := range items {
				if item.LogicalPath == parts.Path {
					if err := apiClient.RotateSecret(item.ID, value); err != nil {
						return err
					}
					fmt.Fprintf(cmd.OutOrStdout(), "Secret saved: %s\n", parts.Path)
					return nil
				}
			}
			if err := apiClient.CreateSecret(parts.Workspace, parts.Project, parts.Environment, parts.Name, value); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Secret saved: %s\n", parts.Path)
			return nil
		},
	}
	command.Flags().StringVar(&workspace, "workspace", "", "workspace id")
	command.Flags().StringVar(&project, "project", "", "project id")
	command.Flags().StringVar(&env, "env", "", "environment id")
	return command
}

func newSecretsRotateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rotate <id> <new-value>",
		Short: "Create a new active secret version",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			if err := apiClient.RotateSecret(args[0], args[1]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Secret rotated: %s\n", args[0])
			return nil
		},
	}
}

func newSecretsRevokeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <id> <version>",
		Short: "Revoke a secret version",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			version, err := strconv.Atoi(args[1])
			if err != nil {
				return failf("version must be an integer")
			}
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			if err := apiClient.RevokeVersion(args[0], version); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Version %d revoked for secret %s\n", version, args[0])
			return nil
		},
	}
}

type parsedPath struct {
	Workspace   string
	Project     string
	Environment string
	Name        string
	Path        string
}

func parseSecretPath(path string, workspace string, project string, env string) (parsedPath, error) {
	segments := strings.Split(path, "/")
	if len(segments) == 4 {
		return parsedPath{Workspace: segments[0], Project: segments[1], Environment: segments[2], Name: segments[3], Path: path}, nil
	}
	if len(segments) == 1 && workspace != "" && project != "" && env != "" {
		return parsedPath{Workspace: workspace, Project: project, Environment: env, Name: path, Path: strings.Join([]string{workspace, project, env, path}, "/")}, nil
	}
	return parsedPath{}, failf("path must be workspace/project/env/name or provide --workspace --project --env")
}

func newSessionClient() (*client.Client, error) {
	session, err := config.Load()
	if err != nil {
		return nil, err
	}
	return client.New(session.APIURL, session.AccessToken), nil
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return "-"
	}
	return value.UTC().Format(time.RFC3339)
}
