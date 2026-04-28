package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const defaultAPIURL = "http://localhost:8080"

var verbose bool

func Execute() error {
	return NewRootCommand().Execute()
}

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "devsvault",
		Short:         "DevsVault local development CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().BoolVar(&verbose, "verbose", false, "print additional diagnostic information")
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	root.AddCommand(newLoginCommand(), newLogoutCommand(), newSecretsCommand(), newRunCommand())
	return root
}

func failf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
