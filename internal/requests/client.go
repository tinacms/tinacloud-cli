package requests

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

const defaultRequestTimeout = 30 * time.Second

var ErrRequestTimeout = errors.New("request timed out")

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{
		http: &http.Client{
			Timeout: defaultRequestTimeout,
		},
	}
}

func (c *Client) requestError(operation string, err error) error {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return fmt.Errorf("%s: %w after %s", operation, ErrRequestTimeout, c.http.Timeout)
	}

	return fmt.Errorf("%s: %w", operation, err)
}
