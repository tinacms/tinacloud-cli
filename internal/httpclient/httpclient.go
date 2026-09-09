package httpclient

import (
	"errors"
	"net/http"
	"time"
)

var ErrUnauthorized = errors.New("no access")

var Client = http.Client{
	Timeout: 10 * time.Second,
}
