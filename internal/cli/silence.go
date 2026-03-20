package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/local/groundcover-cli/internal/gc"
	"github.com/spf13/cobra"
)

func newSilenceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "silence",
		Short:   "Manage silences",
		Aliases: []string{"silences"},
	}

	cmd.AddCommand(newSilenceListCmd())
	cmd.AddCommand(newSilenceCreateCmd())
	return cmd
}

func newSilenceListCmd() *cobra.Command {
	var active bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List silences",
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

			var activePtr *bool
			if cmd.Flags().Changed("active") {
				activePtr = &active
			}

			silences, err := client.ListSilences(ctx, activePtr)
			if err != nil {
				return err
			}

			switch strings.ToLower(cfg.Output) {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(silences)
			case "table", "":
				tw := table.NewWriter()
				tw.SetOutputMirror(os.Stdout)
				tw.AppendHeader(table.Row{"UUID", "STARTS", "ENDS", "COMMENT", "MATCHERS"})
				for _, s := range silences {
					starts := ""
					ends := ""
					if s.StartsAt != nil {
						starts = s.StartsAt.UTC().Format(time.RFC3339)
					}
					if s.EndsAt != nil {
						ends = s.EndsAt.UTC().Format(time.RFC3339)
					}
					tw.AppendRow(table.Row{s.UUID, starts, ends, s.Comment, formatSilenceMatchers(s.Matchers)})
				}
				tw.Render()
				return nil
			default:
				return fmt.Errorf("invalid --output %q (expected table|json)", cfg.Output)
			}
		},
	}

	cmd.Flags().BoolVar(&active, "active", false, "Only show active silences")
	return cmd
}

func newSilenceCreateCmd() *cobra.Command {
	var comment string
	var startsAt string
	var endsAt string
	var duration time.Duration
	var matcherExprs []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a silence",
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

			start, err := parseOptionalRFC3339(startsAt)
			if err != nil {
				return fmt.Errorf("invalid --starts-at: %w", err)
			}
			if start == nil {
				n := time.Now().UTC()
				start = &n
			}

			end, err := parseOptionalRFC3339(endsAt)
			if err != nil {
				return fmt.Errorf("invalid --ends-at: %w", err)
			}
			if end == nil {
				if duration <= 0 {
					duration = 1 * time.Hour
				}
				n := start.Add(duration)
				end = &n
			}
			if !end.After(*start) {
				return fmt.Errorf("ends-at must be after starts-at")
			}

			matchers := make([]gc.SilenceMatcher, 0, len(matcherExprs))
			for _, expr := range matcherExprs {
				m, err := parseSilenceMatcher(expr)
				if err != nil {
					return err
				}
				matchers = append(matchers, m)
			}

			if strings.TrimSpace(comment) == "" {
				comment = "created by gc"
			}

			resp, err := client.CreateSilence(ctx, gc.CreateSilenceRequest{
				StartsAt: start,
				EndsAt:   end,
				Comment:  comment,
				Matchers: matchers,
			})
			if err != nil {
				return err
			}

			switch strings.ToLower(cfg.Output) {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(resp)
			case "table", "":
				tw := table.NewWriter()
				tw.SetOutputMirror(os.Stdout)
				tw.AppendHeader(table.Row{"UUID", "STARTS", "ENDS", "COMMENT", "MATCHERS"})
				starts := ""
				ends := ""
				if resp.StartsAt != nil {
					starts = resp.StartsAt.UTC().Format(time.RFC3339)
				}
				if resp.EndsAt != nil {
					ends = resp.EndsAt.UTC().Format(time.RFC3339)
				}
				tw.AppendRow(table.Row{resp.UUID, starts, ends, resp.Comment, formatSilenceMatchers(resp.Matchers)})
				tw.Render()
				return nil
			default:
				return fmt.Errorf("invalid --output %q (expected table|json)", cfg.Output)
			}
		},
	}

	cmd.Flags().StringVar(&comment, "comment", "", "Silence comment")
	cmd.Flags().StringVar(&startsAt, "starts-at", "", "Silence start time (RFC3339, default now)")
	cmd.Flags().StringVar(&endsAt, "ends-at", "", "Silence end time (RFC3339, default starts-at + duration)")
	cmd.Flags().DurationVar(&duration, "duration", 1*time.Hour, "Silence duration (used when --ends-at not set)")
	cmd.Flags().StringArrayVar(&matcherExprs, "matcher", nil, "Matcher: name=value | name!=value | name~regex | name!~regex (repeatable)")
	return cmd
}

func parseOptionalRFC3339(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		u := t.UTC()
		return &u, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, err
	}
	u := t.UTC()
	return &u, nil
}

func parseSilenceMatcher(expr string) (gc.SilenceMatcher, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return gc.SilenceMatcher{}, fmt.Errorf("empty --matcher")
	}

	var op string
	for _, candidate := range []string{"!~", "~", "!=", "="} {
		if strings.Contains(expr, candidate) {
			op = candidate
			break
		}
	}
	if op == "" {
		return gc.SilenceMatcher{}, fmt.Errorf("invalid --matcher %q (expected name=value | name!=value | name~regex | name!~regex)", expr)
	}

	parts := strings.SplitN(expr, op, 2)
	name := strings.TrimSpace(parts[0])
	value := ""
	if len(parts) == 2 {
		value = strings.TrimSpace(parts[1])
	}
	if name == "" {
		return gc.SilenceMatcher{}, fmt.Errorf("invalid --matcher %q (missing name)", expr)
	}

	isEqual := true
	isRegex := false
	switch op {
	case "=":
		isEqual = true
		isRegex = false
	case "!=":
		isEqual = false
		isRegex = false
	case "~":
		isEqual = true
		isRegex = true
	case "!~":
		isEqual = false
		isRegex = true
	}

	return gc.SilenceMatcher{
		Name:    name,
		Value:   value,
		IsEqual: boolPtr(isEqual),
		IsRegex: boolPtr(isRegex),
	}, nil
}

func boolPtr(b bool) *bool { return &b }

func formatSilenceMatchers(ms []gc.SilenceMatcher) string {
	if len(ms) == 0 {
		return ""
	}
	parts := make([]string, 0, len(ms))
	for _, m := range ms {
		op := "="
		eq := true
		rx := false
		if m.IsEqual != nil {
			eq = *m.IsEqual
		}
		if m.IsRegex != nil {
			rx = *m.IsRegex
		}
		switch {
		case rx && eq:
			op = "~"
		case rx && !eq:
			op = "!~"
		case !rx && !eq:
			op = "!="
		default:
			op = "="
		}
		parts = append(parts, m.Name+op+m.Value)
	}
	return strings.Join(parts, ",")
}
