package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

type SecretMetadata struct {
	ID            string    `json:"id"`
	WorkspaceID   string    `json:"workspace_id"`
	ProjectID     string    `json:"project_id"`
	EnvironmentID string    `json:"environment_id"`
	Name          string    `json:"name"`
	LogicalPath   string    `json:"logical_path"`
	ActiveVersion int       `json:"active_version"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func New(baseURL string, token string) *Client {
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), Token: token, HTTPClient: &http.Client{Timeout: 10 * time.Second}}
}

func (c *Client) Login(subject string, actorType string) (string, time.Time, error) {
	var response struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	request := map[string]string{"subject": subject, "actor_type": actorType}
	if err := c.doJSON(http.MethodPost, "/api/v1/auth/login", request, &response, false); err != nil {
		return "", time.Time{}, err
	}
	if response.AccessToken == "" || response.ExpiresIn <= 0 {
		return "", time.Time{}, errors.New("login response did not include a usable token")
	}
	return response.AccessToken, time.Now().UTC().Add(time.Duration(response.ExpiresIn) * time.Second), nil
}

func (c *Client) ListSecrets() ([]SecretMetadata, error) {
	var response struct {
		Items []SecretMetadata `json:"items"`
	}
	if err := c.doJSON(http.MethodGet, "/api/v1/secrets", nil, &response, true); err != nil {
		return nil, err
	}
	return response.Items, nil
}

func (c *Client) GetSecret(path string) (string, error) {
	var response struct {
		Value string `json:"value"`
	}
	endpoint := "/api/v1/secrets/resolve?path=" + url.QueryEscape(path)
	if err := c.doJSON(http.MethodGet, endpoint, nil, &response, true); err != nil {
		return "", err
	}
	return response.Value, nil
}

func (c *Client) CreateSecret(workspace string, project string, env string, name string, value string) error {
	request := map[string]string{"workspace_id": workspace, "project_id": project, "environment_id": env, "name": name, "value": value}
	return c.doJSON(http.MethodPost, "/api/v1/secrets", request, nil, true)
}

func (c *Client) RotateSecret(id string, value string) error {
	request := map[string]string{"value": value}
	return c.doJSON(http.MethodPost, "/api/v1/secrets/"+url.PathEscape(id)+"/versions", request, nil, true)
}

func (c *Client) RevokeVersion(id string, version int) error {
	endpoint := fmt.Sprintf("/api/v1/secrets/%s/versions/%d/revoke", url.PathEscape(id), version)
	return c.doJSON(http.MethodPost, endpoint, nil, nil, true)
}

func (c *Client) doJSON(method string, endpoint string, body any, target any, auth bool) error {
	var reader *bytes.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(payload)
	} else {
		reader = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, c.BaseURL+endpoint, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if auth {
		if c.Token == "" {
			return errors.New("missing access token")
		}
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s %s failed with status %d", method, endpointPath(endpoint), resp.StatusCode)
	}
	if target == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func endpointPath(endpoint string) string {
	if index := strings.Index(endpoint, "?"); index >= 0 {
		return endpoint[:index]
	}
	return endpoint
}
