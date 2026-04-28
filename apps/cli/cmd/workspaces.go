package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newWorkspacesCommand() *cobra.Command {
	command := &cobra.Command{Use: "workspaces", Short: "Manage workspaces"}
	command.AddCommand(newWorkspacesCreateCommand(), newWorkspacesListCommand(), newWorkspacesGetCommand(), newWorkspacesUpdateCommand(), newWorkspacesDeleteCommand())
	return command
}

func newWorkspacesCreateCommand() *cobra.Command {
	var description string
	command := &cobra.Command{
		Use:   "create <name> <slug>",
		Short: "Create a workspace",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			workspace, err := apiClient.CreateWorkspace(args[0], args[1], description)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Workspace created: %s\n", workspace.ID)
			return nil
		},
	}
	command.Flags().StringVar(&description, "description", "", "workspace description")
	return command
}

func newWorkspacesListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List workspaces",
		RunE: func(cmd *cobra.Command, _ []string) error {
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			items, err := apiClient.ListWorkspaces()
			if err != nil {
				return err
			}
			if len(items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No workspaces found")
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

func newWorkspacesGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Show a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			workspace, err := apiClient.GetWorkspace(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ID\t%s\nSLUG\t%s\nNAME\t%s\nDESCRIPTION\t%s\n", workspace.ID, workspace.Slug, workspace.Name, workspace.Description)
			return nil
		},
	}
}

func newWorkspacesUpdateCommand() *cobra.Command {
	var description string
	command := &cobra.Command{
		Use:   "update <id> <name>",
		Short: "Update a workspace",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			workspace, err := apiClient.UpdateWorkspace(args[0], args[1], description)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Workspace updated: %s\n", workspace.ID)
			return nil
		},
	}
	command.Flags().StringVar(&description, "description", "", "workspace description")
	return command
}

func newWorkspacesDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			if err := apiClient.DeleteWorkspace(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Workspace deleted: %s\n", args[0])
			return nil
		},
	}
}
