package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/local/groundcover-cli/internal/gc"
	"github.com/spf13/cobra"
)

type rootFlags struct {
	APIKey    string
	BackendID string
	BaseURL   string
	Timeout   time.Duration
	Output    string
	Config    string
}

var rf rootFlags

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "gc",
		Short:         "Groundcover CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&rf.APIKey, "api-key", "", "Groundcover API key (env: GC_API_KEY, GROUNDCOVER_API_KEY)")
	cmd.PersistentFlags().StringVar(&rf.BackendID, "backend-id", "", "Groundcover backend ID (env: GC_BACKEND_ID, GROUNDCOVER_BACKEND_ID)")
	cmd.PersistentFlags().StringVar(&rf.BaseURL, "base-url", "", "Groundcover API base URL (env: GC_BASE_URL, GROUNDCOVER_API_URL)")
	cmd.PersistentFlags().DurationVar(&rf.Timeout, "timeout", 30*time.Second, "Request timeout")
	cmd.PersistentFlags().StringVar(&rf.Output, "output", "table", "Output format: table|json")
	cmd.PersistentFlags().StringVar(&rf.Config, "config", "", "Path to config file (default: ~/.config/groundcover-cli/config.json)")

	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newSilenceCmd())

	return cmd
}

func Execute() error {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

func resolvedConfig(requireAuth bool) (gc.Config, error) {
	cfg := gc.Config{
		APIKey:     gc.FirstNonEmpty(rf.APIKey, os.Getenv("GC_API_KEY"), os.Getenv("GROUNDCOVER_API_KEY")),
		BackendID:  gc.FirstNonEmpty(rf.BackendID, os.Getenv("GC_BACKEND_ID"), os.Getenv("GROUNDCOVER_BACKEND_ID")),
		BaseURL:    gc.FirstNonEmpty(rf.BaseURL, os.Getenv("GC_BASE_URL"), os.Getenv("GROUNDCOVER_API_URL")),
		Timeout:    rf.Timeout,
		Output:     rf.Output,
		ConfigFile: strings.TrimSpace(rf.Config),
	}

	if cfg.ConfigFile == "" {
		p, err := gc.DefaultConfigPath()
		if err != nil {
			return gc.Config{}, err
		}
		cfg.ConfigFile = p
	}

	dc, err := gc.LoadDiskConfig(cfg.ConfigFile)
	configMissing := os.IsNotExist(err)
	if err == nil {
		cfg = gc.ApplyDiskConfig(cfg, dc)
	} else if !configMissing {
		return gc.Config{}, err
	}

	setupConfig := func(firstTime bool) error {
		msg := "Config needs setup"
		if firstTime {
			msg = "First-time setup"
		}
		fmt.Fprintf(os.Stderr, "%s: writing config to %s\n", msg, cfg.ConfigFile)

		backend, err := gc.PromptString("Backend ID", cfg.BackendID, true)
		if err != nil {
			return err
		}
		apiKey, err := gc.PromptSecret("API key", cfg.APIKey, true)
		if err != nil {
			return err
		}

		baseDefault := cfg.BaseURL
		if strings.TrimSpace(baseDefault) == "" {
			baseDefault = "https://api.groundcover.com"
		}
		baseURL, err := gc.PromptString("Base URL", baseDefault, false)
		if err != nil {
			return err
		}

		newDC := gc.DiskConfig{APIKey: apiKey, BackendID: backend, BaseURL: baseURL}
		if err := gc.WriteDiskConfig(cfg.ConfigFile, newDC); err != nil {
			return err
		}
		cfg = gc.ApplyDiskConfig(cfg, newDC)
		return nil
	}

	if requireAuth && configMissing {
		if gc.IsInteractiveStdin() {
			if err := setupConfig(true); err != nil {
				return gc.Config{}, err
			}
		} else {
			if strings.TrimSpace(cfg.APIKey) == "" || strings.TrimSpace(cfg.BackendID) == "" {
				return gc.Config{}, fmt.Errorf("config not found at %s and stdin is non-interactive; set GC_API_KEY/GC_BACKEND_ID or run in a terminal to initialize config", cfg.ConfigFile)
			}
			baseURL := cfg.BaseURL
			if strings.TrimSpace(baseURL) == "" {
				baseURL = "https://api.groundcover.com"
			}
			autoDC := gc.DiskConfig{APIKey: cfg.APIKey, BackendID: cfg.BackendID, BaseURL: baseURL}
			if err := gc.WriteDiskConfig(cfg.ConfigFile, autoDC); err != nil {
				return gc.Config{}, err
			}
			cfg = gc.ApplyDiskConfig(cfg, autoDC)
		}
	}

	if requireAuth && (strings.TrimSpace(cfg.APIKey) == "" || strings.TrimSpace(cfg.BackendID) == "") {
		if !gc.IsInteractiveStdin() {
			return gc.Config{}, fmt.Errorf("config at %s is missing required auth fields and stdin is non-interactive; set --api-key/--backend-id or update config", cfg.ConfigFile)
		}
		if err := setupConfig(false); err != nil {
			return gc.Config{}, err
		}
	}

	return cfg, nil
}
