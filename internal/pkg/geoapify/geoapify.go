package geoapify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Feature struct {
	Properties struct {
		Formatted  string  `json:"formatted"`
		Lat        float64 `json:"lat"`
		Lon        float64 `json:"lon"`
		Name       string  `json:"name"`
		ResultType string  `json:"result_type"`
	} `json:"properties"`
	Type string `json:"type"`
}

type Client struct {
	url    *url.URL
	apiKey string
}

func NewClient(serverURL, apiKey string) (Client, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return Client{}, fmt.Errorf("failed to parse server url: %w", err)
	}

	return Client{url: u, apiKey: apiKey}, nil
}

func (c *Client) GeocodingSearch(ctx context.Context, query string) ([]Feature, error) {
	queryURL, err := url.Parse(c.url.String())
	if err != nil {
		return []Feature{}, fmt.Errorf("failed to parse base url: %w", err)
	}
	values := queryURL.Query()
	values.Set("apiKey", c.apiKey)
	values.Set("text", query)

	queryURL.RawQuery = values.Encode()
	queryURL.Path = "/v1/geocode/search"

	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL.String(), nil)
	if err != nil {
		return []Feature{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return []Feature{}, fmt.Errorf("failed to get features from API: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []Feature{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return []Feature{}, fmt.Errorf("error requesting features: %q", string(body))
	}

	response := struct {
		Features []Feature `json:"features"`
	}{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return []Feature{}, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return response.Features, nil
}
