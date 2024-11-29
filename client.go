// Copyright (c) 2020, 2023, Geert JM Vanderkelen

package gogql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Result struct {
	Data   json.RawMessage  `json:"data,omitempty"`
	Errors []gqlerror.Error `json:"errors,omitempty"`
}

// Client defines a GraphQL client connecting to the endpoint using
// HTTP.
type Client struct {
	ContentType string
	HttpClient  *http.Client

	apiEndpoint string
	headers     map[string]string
}

// Payload defines what we send to the GraphQL endpoint.
type Payload struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// NewClient returns a new GraphQL Client connecting to the apiEndpoint.
func NewClient(apiEndpoint string) *Client {
	return &Client{
		ContentType: "application/json",
		apiEndpoint: apiEndpoint,
		HttpClient:  http.DefaultClient,
	}
}

// Execute executes the GraphQL query.
func (c *Client) Execute(query string, result any) []error {
	return c.execute(query, result, nil)
}

func (c *Client) ExecuteWithVars(query string, result interface{}, vars map[string]any) []error {
	return c.execute(query, result, vars)
}

func (c *Client) SetHeader(key, value string) {
	if c.headers == nil {
		c.headers = map[string]string{}
	}

	c.headers[key] = value
}

func (c *Client) execute(query string, result interface{}, vars map[string]any) []error {
	payload := &Payload{
		Query:     query,
		Variables: vars,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return []error{fmt.Errorf("encoding request (%w)", err)}
	}

	req, err := http.NewRequest(http.MethodPost, c.apiEndpoint, bytes.NewReader(body))
	if err != nil {
		return []error{fmt.Errorf("creating HTTP request (%w)", err)}
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	if c.ContentType != "" {
		req.Header.Set("Content-Type", c.ContentType)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return []error{fmt.Errorf("gogql: sending request (%w)", err)}
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return []error{fmt.Errorf("gogql: reading response body (%w)", err)}
	}
	defer func() { _ = resp.Body.Close() }()

	var data Result
	if err := json.Unmarshal(buf, &data); err != nil {
		return []error{fmt.Errorf("gogql: unmarshalling data (%w)", err)}
	}

	if len(data.Errors) > 0 {
		var errs []error
		for _, e := range data.Errors {
			errs = append(errs, &e)
		}
		return errs
	}

	if err := json.Unmarshal(data.Data, result); err != nil {
		return []error{fmt.Errorf("gogql: unmarshalling data (%w)", err)}
	}
	return nil
}
