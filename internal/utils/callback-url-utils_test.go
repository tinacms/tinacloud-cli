package utils

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"testing"
	"time"
)

func TestWaitForCallback(t *testing.T) {
	ln, _, err := CallbackListener()
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	if !ln.Addr().(*net.TCPAddr).IP.IsLoopback() {
		t.Fatalf("listener address = %s, want loopback", ln.Addr())
	}
	result := make(chan error, 1)
	go func() {
		values, err := waitForCallback(ln, "/callback", func(v url.Values) bool { return v.Get("code") != "" }, 2*time.Second, 50*time.Millisecond)
		if err == nil && values.Get("code") != "valid" {
			err = fmt.Errorf("code = %q, want valid", values.Get("code"))
		}
		result <- err
	}()
	stalled, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer stalled.Close()
	for _, request := range []string{"GET /wrong?code=x HTTP/1.1\r\nHost: localhost\r\n\r\n", "GET /callback HTTP/1.1\r\nHost: localhost\r\n\r\n", "GET /callback?code=valid HTTP/1.1\r\nHost: localhost\r\n\r\n"} {
		conn, err := net.Dial("tcp", ln.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprint(conn, request)
		io.Copy(io.Discard, conn)
		_ = conn.Close()
	}
	if err := <-result; err != nil {
		t.Fatal(err)
	}
}

func TestWaitForCallbackTimeout(t *testing.T) {
	ln, _, err := CallbackListener()
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	_, err = waitForCallback(ln, "/", func(url.Values) bool { return true }, 50*time.Millisecond, 10*time.Millisecond)
	if err == nil {
		t.Fatal("waitForCallback() = nil error, want timeout")
	}
}
