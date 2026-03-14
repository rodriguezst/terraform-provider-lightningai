package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://lightning.ai/v1"

// Client is the Lightning AI API client.
type Client struct {
	apiKey     string
	userID     string
	projectID  string
	httpClient *http.Client
}

// NewClient creates a new Lightning AI API client.
func NewClient(apiKey, userID, projectID string) *Client {
	return &Client{
		apiKey:    apiKey,
		userID:    userID,
		projectID: projectID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Studio represents a Lightning AI Studio (cloudspace).
type Studio struct {
	ID          string            `json:"id,omitempty"`
	Name        string            `json:"name"`
	ClusterID   string            `json:"cluster_id,omitempty"`
	CodeStatus  *StudioCodeStatus `json:"-"`
}

// StudioCodeStatus represents the status of a studio.
type StudioCodeStatus struct {
	Phase    string `json:"phase"`
	PublicIP string `json:"public_ip,omitempty"`
	SSHHost  string `json:"ssh_host,omitempty"`
	SSHUsername string `json:"ssh_username,omitempty"`
}

// CreateStudioRequest is the request body for creating a studio.
type CreateStudioRequest struct {
	Name string `json:"name"`
}

// StartStudioRequest is the request body for starting a studio.
type StartStudioRequest struct {
	MachineType   string `json:"machine_type,omitempty"`
	Interruptible bool   `json:"interruptible,omitempty"`
	ClusterID     string `json:"cluster_id,omitempty"`
}

// ExecuteRequest is the request body for executing a command in a studio.
type ExecuteRequest struct {
	Command string `json:"command"`
}

// ExecuteResponse is the response body from executing a command.
type ExecuteResponse struct {
	Output   string `json:"output,omitempty"`
	ExitCode int    `json:"exit_code,omitempty"`
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(data)
	}

	url := fmt.Sprintf("%s%s", baseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func readBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// CreateStudio creates a new studio.
func (c *Client) CreateStudio(ctx context.Context, name string) (*Studio, error) {
	path := fmt.Sprintf("/projects/%s/cloudspaces", c.projectID)
	reqBody := CreateStudioRequest{Name: name}

	resp, err := c.doRequest(ctx, http.MethodPost, path, reqBody)
	if err != nil {
		return nil, err
	}
	data, err := readBody(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("create studio failed (status %d): %s", resp.StatusCode, string(data))
	}

	var studio Studio
	if err := json.Unmarshal(data, &studio); err != nil {
		return nil, fmt.Errorf("failed to parse create studio response: %w", err)
	}
	return &studio, nil
}

// GetStudio retrieves a studio by ID.
func (c *Client) GetStudio(ctx context.Context, studioID string) (*Studio, error) {
	path := fmt.Sprintf("/projects/%s/cloudspaces?userId=%s", c.projectID, c.userID)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	data, err := readBody(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("list studios failed (status %d): %s", resp.StatusCode, string(data))
	}

	var result struct {
		Cloudspaces []Studio `json:"cloudspaces"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse list studios response: %w", err)
	}

	for _, s := range result.Cloudspaces {
		if s.ID == studioID {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("studio %s not found", studioID)
}

// GetStudioStatus retrieves the code status of a studio.
func (c *Client) GetStudioStatus(ctx context.Context, studioID string) (*StudioCodeStatus, error) {
	path := fmt.Sprintf("/projects/%s/cloudspaces/%s/codestatus", c.projectID, studioID)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	data, err := readBody(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("get studio status failed (status %d): %s", resp.StatusCode, string(data))
	}

	var status StudioCodeStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("failed to parse studio status response: %w", err)
	}
	return &status, nil
}

// StartStudio starts a studio.
func (c *Client) StartStudio(ctx context.Context, studioID, machineType string, interruptible bool) error {
	path := fmt.Sprintf("/projects/%s/cloudspaces/%s/start", c.projectID, studioID)
	reqBody := StartStudioRequest{
		MachineType:   machineType,
		Interruptible: interruptible,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, path, reqBody)
	if err != nil {
		return err
	}
	data, err := readBody(resp)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("start studio failed (status %d): %s", resp.StatusCode, string(data))
	}
	return nil
}

// StopStudio stops a studio.
func (c *Client) StopStudio(ctx context.Context, studioID string) error {
	path := fmt.Sprintf("/projects/%s/cloudspaces/%s/stop", c.projectID, studioID)

	resp, err := c.doRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	data, err := readBody(resp)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("stop studio failed (status %d): %s", resp.StatusCode, string(data))
	}
	return nil
}

// DeleteStudio deletes a studio.
func (c *Client) DeleteStudio(ctx context.Context, studioID string) error {
	path := fmt.Sprintf("/projects/%s/cloudspaces/%s", c.projectID, studioID)

	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	data, err := readBody(resp)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("delete studio failed (status %d): %s", resp.StatusCode, string(data))
	}
	return nil
}

// ExecuteCommand executes a command in a studio.
func (c *Client) ExecuteCommand(ctx context.Context, studioID, command string) (*ExecuteResponse, error) {
	path := fmt.Sprintf("/projects/%s/cloudspaces/%s/execute", c.projectID, studioID)
	reqBody := ExecuteRequest{Command: command}

	resp, err := c.doRequest(ctx, http.MethodPost, path, reqBody)
	if err != nil {
		return nil, err
	}
	data, err := readBody(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("execute command failed (status %d): %s", resp.StatusCode, string(data))
	}

	var result ExecuteResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse execute response: %w", err)
	}
	return &result, nil
}

// ProjectID returns the configured project ID.
func (c *Client) ProjectID() string {
	return c.projectID
}

// UserID returns the configured user ID.
func (c *Client) UserID() string {
	return c.userID
}
