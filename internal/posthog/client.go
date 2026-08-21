package posthog

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	sysinfo "tina.io/tinacloud-cli/internal/sys-info"
)

const (
	libName    = "tinacloud-cli-posthog"
	libVersion = "1.0.0"

	// requestTimeout must stay below closeTimeout so a request always finishes
	// (or fails) in time for Close to observe it, instead of Close abandoning
	// a request that was still going to complete.
	requestTimeout = 4 * time.Second
	closeTimeout   = 5 * time.Second
)

type Client struct {
	apiKey   string
	endpoint string
	http     *http.Client
	wg       sync.WaitGroup
}

func NewClient(apiKey, endpoint string) *Client {
	return &Client{
		apiKey:   apiKey,
		endpoint: endpoint,
		http:     &http.Client{Timeout: requestTimeout},
	}
}

// Enqueue sends a capture event in the background.
func (c *Client) Enqueue(event Capture) {
	c.send(map[string]any{
		"event":       event.Event,
		"distinct_id": event.DistinctId,
		"properties":  event.Properties,
	})
}

// Alias links distinctID to alias via a $create_alias event.
func (c *Client) Alias(distinctID, alias string) {
	c.send(map[string]any{
		"event": "$create_alias",
		"properties": map[string]any{
			"distinct_id": distinctID,
			"alias":       alias,
		},
	})
}

func (c *Client) send(event map[string]any) {
	props, _ := event["properties"].(map[string]any)
	if props == nil {
		props = map[string]any{}
	}
	props["$lib"] = libName
	props["$lib_version"] = libVersion
	for k, v := range sysinfo.GetSystemContext().ToProperties() {
		props[k] = v
	}
	event["properties"] = props

	body, err := json.Marshal(map[string]any{
		"api_key": c.apiKey,
		"batch":   []map[string]any{event},
	})
	if err != nil {
		return
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		req, err := http.NewRequest(http.MethodPost, c.endpoint+"/batch/", bytes.NewReader(body))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			return
		}
		_ = resp.Body.Close()
	}()
}

func (c *Client) Close() {
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(closeTimeout):
	}
}
