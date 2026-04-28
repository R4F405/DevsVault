package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newEnvironmentsCommand() *cobra.Command {
	command := &cobra.Command{Use: "environments", Short: "Manage environments"}
	command.AddCommand(newEnvironmentsCreateCommand(), newEnvironmentsListCommand(), newEnvironmentsGetCommand(), newEnvironmentsDeleteCommand())
	return command
}

func newEnvironmentsCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create <project-id> <name> <slug>",
		Short: "Create an environment",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			environment, err := apiClient.CreateEnvironment(args[0], args[1], args[2])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Environment created: %s\n", environment.ID)
			return nil
		},
	}
}

func newEnvironmentsListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <project-id>",
		Short: "List environments in a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			items, err := apiClient.ListEnvironments(args[0])
			if err != nil {
				return err
			}
			if len(items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No environments found")
				return nil
			}
			writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(writer, "ID\tSLUG\tNAME\tUPDATED")
			for _, item := range items {
				fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n", item.ID, item.Slug, item.Name, formatTime(item.UpdatedAt))
			}
			return writer.Flush()
		},
	}
}

func newEnvironmentsGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Show an environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			environment, err := apiClient.GetEnvironment(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ID\t%s\nPROJECT\t%s\nSLUG\t%s\nNAME\t%s\n", environment.ID, environment.ProjectID, environment.Slug, environment.Name)
			return nil
		},
	}
}

func newEnvironmentsDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			environment, err := apiClient.GetEnvironment(args[0])
			if err != nil {
				return err
			}
			if err := apiClient.DeleteEnvironment(environment.ProjectID, args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Environment deleted: %s\n", args[0])
			return nil
		},
	}
}
