package cli

import "github.com/spf13/cobra"

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List resources",
		Aliases: []string{"ls"},
	}

	cmd.AddCommand(newMonitorsCmd())
	cmd.AddCommand(newMonitorIssuesCmd())
	return cmd
}
