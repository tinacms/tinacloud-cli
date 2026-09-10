package httpclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClassifyError(t *testing.T) {
	if got := ClassifyHTTPError(context.DeadlineExceeded); !errors.Is(got, ErrTimeout) {
		t.Errorf("ClassifyError(context.DeadlineExceeded) = %v, want ErrTimeout", got)
	}

	want := errors.New("connection failed")
	if got := ClassifyHTTPError(want); !errors.Is(got, want) {
		t.Errorf("ClassifyError() = %v, want %v", got, want)
	}
}

func TestClientTimeout(t *testing.T) {
	originalTimeout := Client.Timeout
	Client.Timeout = 10 * time.Millisecond
	t.Cleanup(func() {
		Client.Timeout = originalTimeout
	})

	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		time.Sleep(50 * time.Millisecond)
	}))
	t.Cleanup(server.Close)

	resp, err := Client.Get(server.URL)
	if resp != nil {
		resp.Body.Close()
	}
	if !errors.Is(ClassifyHTTPError(err), ErrTimeout) {
		t.Errorf("Client.Get() = %v, want ErrTimeout", err)
	}
}
