package cli

import "github.com/spf13/cobra"

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get",
		Short:   "Get resources",
		Aliases: []string{"g"},
	}

	cmd.AddCommand(newMonitorsCmd())
	return cmd
}
