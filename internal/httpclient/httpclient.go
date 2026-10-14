package httpclient

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

var ErrUnauthorized = errors.New("authorisation required, please run 'tinacloud auth login'")
var ErrTimeout = errors.New("request timed out; please try again")

const requestTimeout = 15 * time.Second

var userAgent = "tinacloud-cli/undefined"
var version = "undefined"

var Client = NewClient()

func SetVersion(value string) {
	if value == "" {
		return
	}
	version = value
	userAgent = "tinacloud-cli/" + value
}

func Version() string {
	return version
}

func NewClient() *http.Client {
	return &http.Client{
		Timeout:   requestTimeout,
		Transport: userAgentTransport{next: http.DefaultTransport},
	}
}

type userAgentTransport struct {
	next http.RoundTripper
}

func (t userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("User-Agent", userAgent)
	return t.next.RoundTrip(req)
}

func ClassifyHTTPError(err error) error {
	if err == nil {
		return nil
	}

	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return fmt.Errorf("%w: %v", ErrTimeout, err)
	}

	return err
}
