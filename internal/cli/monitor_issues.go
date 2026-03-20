package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/local/groundcover-cli/internal/gc"
	"github.com/spf13/cobra"
)

func newMonitorIssuesCmd() *cobra.Command {
	var envs []string
	var clusters []string
	var namespaces []string
	var workloads []string
	var monitorIDs []string
	var silenced string
	var limit int64
	var skip int64

	cmd := &cobra.Command{
		Use:     "monitor-issues",
		Short:   "List monitor issues",
		Aliases: []string{"monitor-issue", "issues"},
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

			var silencedPtr *bool
			s := strings.ToLower(strings.TrimSpace(silenced))
			switch s {
			case "", "any":
				// nil
			case "true", "t", "1", "yes", "y":
				v := true
				silencedPtr = &v
			case "false", "f", "0", "no", "n":
				v := false
				silencedPtr = &v
			default:
				return fmt.Errorf("invalid --silenced %q (expected true|false|any)", silenced)
			}

			var limitPtr *int64
			var skipPtr *int64
			if cmd.Flags().Changed("limit") {
				limitPtr = &limit
			}
			if cmd.Flags().Changed("skip") {
				skipPtr = &skip
			}

			issues, err := client.ListMonitorIssues(ctx, gc.MonitorIssuesListRequest{
				Envs:       envs,
				Clusters:   clusters,
				Namespaces: namespaces,
				Workloads:  workloads,
				MonitorIDs: monitorIDs,
				Silenced:   silencedPtr,
				Limit:      limitPtr,
				Skip:       skipPtr,
				SortOrder:  "desc",
			})
			if err != nil {
				return err
			}

			switch strings.ToLower(cfg.Output) {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(issues)
			case "table", "":
				tw := table.NewWriter()
				tw.SetOutputMirror(os.Stdout)
				tw.AppendHeader(table.Row{"ISSUE_ID", "STATUS", "SILENCED", "MONITOR_ID", "WORKLOAD", "NAMESPACE", "CLUSTER", "ENV"})
				for _, it := range issues {
					tw.AppendRow(table.Row{
						getString(it, "issueId", "id"),
						getString(it, "status", "state"),
						getBoolString(it, "silenced"),
						getString(it, "monitorId", "monitorUuid", "monitorUUID", "monitor"),
						getString(it, "workload"),
						getString(it, "namespace"),
						getString(it, "cluster"),
						getString(it, "env"),
					})
				}
				tw.Render()
				return nil
			default:
				return fmt.Errorf("invalid --output %q (expected table|json)", cfg.Output)
			}
		},
	}

	cmd.Flags().StringArrayVar(&envs, "env", nil, "Filter by environment (repeatable)")
	cmd.Flags().StringArrayVar(&clusters, "cluster", nil, "Filter by cluster (repeatable)")
	cmd.Flags().StringArrayVar(&namespaces, "namespace", nil, "Filter by namespace (repeatable)")
	cmd.Flags().StringArrayVar(&workloads, "workload", nil, "Filter by workload (repeatable)")
	cmd.Flags().StringArrayVar(&monitorIDs, "monitor-id", nil, "Filter by monitor UUID (repeatable)")
	cmd.Flags().StringVar(&silenced, "silenced", "any", "Filter by silenced: true|false|any")
	cmd.Flags().Int64Var(&limit, "limit", 0, "Max issues to return")
	cmd.Flags().Int64Var(&skip, "skip", 0, "Items to skip (pagination)")
	return cmd
}

func getString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		v, ok := m[k]
		if !ok || v == nil {
			continue
		}
		switch t := v.(type) {
		case string:
			if strings.TrimSpace(t) != "" {
				return t
			}
		case fmt.Stringer:
			return t.String()
		case float64:
			return strconv.FormatInt(int64(t), 10)
		case bool:
			if t {
				return "true"
			}
			return "false"
		}
	}
	return ""
}

func getBoolString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if b, ok := v.(bool); ok {
		if b {
			return "true"
		}
		return "false"
	}
	if s, ok := v.(string); ok {
		return strings.ToLower(strings.TrimSpace(s))
	}
	return ""
}
