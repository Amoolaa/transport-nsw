package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type Client[T any] struct {
	httpClient http.Client
	baseURL    url.URL
}

func New[T any](httpClient http.Client) *Client[T] {
	return &Client[T]{
		httpClient: httpClient,
		baseURL: url.URL{
			Scheme: "https",
			Host:   "api.transport.nsw.gov.au",
			Path:   "/v1",
		},
	}
}

func (c *Client[T]) Do(
	ctx context.Context,
	method string,
	endpoint string,
	queryParams url.Values,
	requestBody any,
) (*T, error) {
	fullURL, err := url.Parse(c.baseURL.String() + endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to join url paths: %w", err)
	}

	if len(queryParams) > 0 {
		fullURL.RawQuery = queryParams.Encode()
	}

	var body io.Reader
	if requestBody != nil {
		data, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("apikey %s", os.Getenv("API_TOKEN")))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}
