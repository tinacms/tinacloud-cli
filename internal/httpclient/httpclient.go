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

var Client = http.Client{
	Timeout: 30 * time.Second,
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
