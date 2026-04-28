package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var execCommand = exec.Command

func newRunCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run -- <command> [args...]",
		Short: "Run a command with DevsVault secrets injected",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := newSessionClient()
			if err != nil {
				return err
			}
			items, err := apiClient.ListSecrets()
			if err != nil {
				return err
			}
			env := os.Environ()
			for _, item := range items {
				value, err := apiClient.GetSecret(item.LogicalPath)
				if err != nil {
					return err
				}
				name := envName(item.Name)
				env = append(env, name+"="+value)
				if verbose {
					fmt.Fprintln(cmd.ErrOrStderr(), "WARNING: secret values visible in output")
					fmt.Fprintf(cmd.ErrOrStderr(), "%s=%s\n", name, value)
				}
			}
			process := execCommand(args[0], args[1:]...)
			process.Env = env
			process.Stdin = cmd.InOrStdin()
			process.Stdout = cmd.OutOrStdout()
			process.Stderr = cmd.ErrOrStderr()
			return process.Run()
		},
	}
}

func envName(name string) string {
	name = strings.ToUpper(name)
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}
