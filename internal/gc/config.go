package gc

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Config struct {
	APIKey     string
	BackendID  string
	BaseURL    string
	Timeout    time.Duration
	Output     string
	ConfigFile string
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.APIKey) == "" {
		return errors.New("missing API key (set --api-key or env GC_API_KEY / GROUNDCOVER_API_KEY)")
	}
	if strings.TrimSpace(c.BackendID) == "" {
		return errors.New("missing backend ID (set --backend-id or env GC_BACKEND_ID / GROUNDCOVER_BACKEND_ID)")
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("invalid timeout: %s", c.Timeout)
	}
	return nil
}

func FirstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
