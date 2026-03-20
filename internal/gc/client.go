package gc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.groundcover.com"

type Client struct {
	baseURL   string
	apiKey    string
	backendID string
	hc        *http.Client
}

type MonitorListItem struct {
	Title string `json:"title"`
	Type  string `json:"type"`
	UUID  string `json:"uuid"`
}

type MonitorListResponse struct {
	Monitors []MonitorListItem `json:"monitors"`
}

type SilenceMatcher struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	IsEqual *bool  `json:"isEqual,omitempty"`
	IsRegex *bool  `json:"isRegex,omitempty"`
}

type Silence struct {
	UUID     string           `json:"uuid"`
	StartsAt *time.Time       `json:"startsAt,omitempty"`
	EndsAt   *time.Time       `json:"endsAt,omitempty"`
	Comment  string           `json:"comment,omitempty"`
	Matchers []SilenceMatcher `json:"matchers,omitempty"`
}

type CreateSilenceRequest struct {
	StartsAt *time.Time       `json:"startsAt,omitempty"`
	EndsAt   *time.Time       `json:"endsAt,omitempty"`
	Comment  string           `json:"comment,omitempty"`
	Matchers []SilenceMatcher `json:"matchers,omitempty"`
}

type MonitorIssuesListRequest struct {
	Clusters   []string `json:"clusters,omitempty"`
	Envs       []string `json:"envs,omitempty"`
	Namespaces []string `json:"namespaces,omitempty"`
	Workloads  []string `json:"workloads,omitempty"`
	MonitorIDs []string `json:"monitorIds,omitempty"`
	Silenced   *bool    `json:"silenced,omitempty"`
	Limit      *int64   `json:"limit,omitempty"`
	Skip       *int64   `json:"skip,omitempty"`
	SortOrder  string   `json:"sortOrder,omitempty"`
}

func NewClient(cfg Config) (*Client, error) {
	base := strings.TrimSpace(cfg.BaseURL)
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("invalid base url %q: %w", base, err)
	}
	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("invalid base url %q: missing host", base)
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")

	return &Client{
		baseURL:   parsed.String(),
		apiKey:    cfg.APIKey,
		backendID: cfg.BackendID,
		hc:        &http.Client{},
	}, nil
}

func (c *Client) doJSON(ctx context.Context, method string, path string, query url.Values, reqBody any, respBody any) error {
	target := c.baseURL + path
	if len(query) > 0 {
		if strings.Contains(target, "?") {
			target += "&" + query.Encode()
		} else {
			target += "?" + query.Encode()
		}
	}

	var bodyReader io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, target, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("X-Backend-Id", c.backendID)
	req.Header.Set("User-Agent", "groundcover-cli")
	req.Header.Set("Accept", "application/json")
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(respBytes, &apiErr)
		msg := strings.TrimSpace(apiErr.Message)
		if msg == "" {
			msg = strings.TrimSpace(string(respBytes))
		}
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return fmt.Errorf("groundcover api error (%d): %s", resp.StatusCode, msg)
	}

	if respBody == nil {
		return nil
	}
	if err := json.Unmarshal(respBytes, respBody); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c *Client) ListMonitors(ctx context.Context, limit *int64) (*MonitorListResponse, error) {
	body := map[string]any{}
	if limit != nil && *limit > 0 {
		body["limit"] = *limit
	}
	var out MonitorListResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/monitors/list", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ListSilences(ctx context.Context, active *bool) ([]Silence, error) {
	q := url.Values{}
	if active != nil {
		if *active {
			q.Set("active", "true")
		} else {
			q.Set("active", "false")
		}
	}
	var out []Silence
	if err := c.doJSON(ctx, http.MethodGet, "/api/monitors/silences", q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateSilence(ctx context.Context, req CreateSilenceRequest) (*Silence, error) {
	var out Silence
	if err := c.doJSON(ctx, http.MethodPost, "/api/monitors/silences", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ListMonitorIssues(ctx context.Context, req MonitorIssuesListRequest) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.doJSON(ctx, http.MethodPost, "/api/monitors/issues/list", nil, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}
