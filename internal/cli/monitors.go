package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/local/groundcover-cli/internal/gc"
	"github.com/spf13/cobra"
)

func newMonitorsCmd() *cobra.Command {
	var limit int64

	cmd := &cobra.Command{
		Use:     "monitors",
		Short:   "List monitors",
		Aliases: []string{"monitor"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := resolvedConfig(true)
			if err != nil {
				return err
			}
			if err := cfg.Validate(); err != nil {
				return err
			}

			client, err := gc.NewClient(cfg)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), cfg.Timeout)
			defer cancel()

			var reqLimit *int64
			if limit > 0 {
				reqLimit = &limit
			}

			resp, err := client.ListMonitors(ctx, reqLimit)
			if err != nil {
				return err
			}
			if resp == nil {
				return errors.New("empty response from Groundcover")
			}

			switch strings.ToLower(cfg.Output) {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(resp)
			case "table", "":
				tw := table.NewWriter()
				tw.SetOutputMirror(os.Stdout)
				tw.AppendHeader(table.Row{"UUID", "TYPE", "TITLE"})
				for _, m := range resp.Monitors {
					tw.AppendRow(table.Row{m.UUID, m.Type, m.Title})
				}
				tw.Render()
				return nil
			default:
				return fmt.Errorf("invalid --output %q (expected table|json)", cfg.Output)
			}
		},
	}

	cmd.Flags().Int64Var(&limit, "limit", 0, "Max monitors to return")
	return cmd
}
