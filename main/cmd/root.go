package cmd

import (
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	// CLIName is the name of the CLI
	CLIName = "kuai"
)

// NewCommand returns a new instance of an Arena command
func NewCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   CLIName,
		Short: "kuai is the command line interface to KuAI",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	command.AddCommand(NewSubmitCommand())
	command.AddCommand(NewListCommand())
	//command.AddCommand(NewGetCommand())
	//command.AddCommand(NewLogViewerCommand())
	command.AddCommand(NewLogsCommand())
	//command.AddCommand(NewDeleteCommand())
	//command.AddCommand(NewTopCommand())
	//command.AddCommand(NewVersionCmd(CLIName))
	//command.AddCommand(NewDataCommand())
	//command.AddCommand(NewCompletionCommand())

	return command
}
