package parser

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type crawler struct {
	client *http.Client
}

func newCrawler() *crawler {
	return &crawler{
		client: &http.Client{},
	}
}

// Crawl returns webpage's content.
func (c *crawler) Crawl(ctx context.Context, url url.URL) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("faield to create request: %w", err)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to send reqeust: status code is %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if err := response.Body.Close(); err != nil {
		return nil, fmt.Errorf("failed to close response body: %w", err)
	}

	return body, nil
}
