package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newProjectsCommand() *cobra.Command {
	command := &cobra.Command{Use: "projects", Short: "Manage projects"}
	command.AddCommand(newProjectsCreateCommand(), newProjectsListCommand(), newProjectsGetCommand(), newProjectsUpdateCommand(), newProjectsDeleteCommand())
	return command
}

func newProjectsCreateCommand() *cobra.Command {
	var description string
	command := &cobra.Command{
		Use:   "create <workspace-id> <name> <slug>",
		Short: "Create a project",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			project, err := apiClient.CreateProject(args[0], args[1], args[2], description)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Project created: %s\n", project.ID)
			return nil
		},
	}
	command.Flags().StringVar(&description, "description", "", "project description")
	return command
}

func newProjectsListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <workspace-id>",
		Short: "List projects in a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			items, err := apiClient.ListProjects(args[0])
			if err != nil {
				return err
			}
			if len(items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No projects found")
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

func newProjectsGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Show a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			project, err := apiClient.GetProject(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ID\t%s\nWORKSPACE\t%s\nSLUG\t%s\nNAME\t%s\nDESCRIPTION\t%s\n", project.ID, project.WorkspaceID, project.Slug, project.Name, project.Description)
			return nil
		},
	}
}

func newProjectsUpdateCommand() *cobra.Command {
	var description string
	command := &cobra.Command{
		Use:   "update <id> <name>",
		Short: "Update a project",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			existing, err := apiClient.GetProject(args[0])
			if err != nil {
				return err
			}
			project, err := apiClient.UpdateProject(existing.WorkspaceID, args[0], args[1], description)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Project updated: %s\n", project.ID)
			return nil
		},
	}
	command.Flags().StringVar(&description, "description", "", "project description")
	return command
}

func newProjectsDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			project, err := apiClient.GetProject(args[0])
			if err != nil {
				return err
			}
			if err := apiClient.DeleteProject(project.WorkspaceID, args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Project deleted: %s\n", args[0])
			return nil
		},
	}
}
