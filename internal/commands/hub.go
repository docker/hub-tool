package commands

import (
	"context"
	"fmt"

	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"

	"github.com/docker/hub-cli-plugin/internal"
)

type options struct {
	showVersion bool
}

//NewHubCmd returns the main command
func NewHubCmd(_ context.Context, _ command.Cli) *cobra.Command {
	var flags options
	cmd := &cobra.Command{
		Short:       "Docker Hub",
		Long:        `A tool to manage your Docker Hub images`,
		Use:         "hub",
		Annotations: map[string]string{},
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.showVersion {
				return runVersion()
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&flags.showVersion, "version", false, "Display version of the scan plugin")

	return cmd
}

func runVersion() error {
	fmt.Println("Version:   ", internal.Version)
	fmt.Println("Git commit:", internal.GitCommit)
	return nil
}
