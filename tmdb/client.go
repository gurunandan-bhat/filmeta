package tmdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/schema"
)

const (
	BaseURLV3 = "https://api.themoviedb.org/3"
)

type Client struct {
	BaseURL    string
	apiKey     string
	HTTPClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		BaseURL: BaseURLV3,
		apiKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}
}

var encoder = schema.NewEncoder()

func (c *Client) sendRequest(req *http.Request, v any) error {

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		body, err := io.ReadAll(res.Body)
		errStr := string(body)
		if err != nil {
			errStr = err.Error()
		}
		return fmt.Errorf("error with status %s for URL %s: %s", res.Status, req.URL.String(), errStr)
	}

	f, isWriter := v.(io.Writer)
	if isWriter {
		_, err = io.Copy(f, res.Body)
		if err != nil {
			return fmt.Errorf("error copying image: %w", err)
		}
		return nil
	}

	if err = json.NewDecoder(res.Body).Decode(v); err != nil {
		return fmt.Errorf("error unmarshaling body: %w", err)
	}

	return nil
}
