package httpclient

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

var ErrUnauthorized = errors.New("no access")
var ErrTimeout = errors.New("request timed out; please try again")

var Client = http.Client{
	Timeout: 30 * time.Second,
}

func ClassifyHTTPError(err error) error {
	if err == nil {
		return nil
	}

	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return ErrTimeout
	}

	return err
}
