package utils

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
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
		values, err := waitForCallback(ln, "/callback", func(v url.Values) (string, string) {
			if v.Get("code") != "" {
				return "Almost there!", "Return to your terminal."
			}
			return "", ""
		}, 2*time.Second, 50*time.Millisecond)
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

func TestWaitForCallbackResponse(t *testing.T) {
	for _, tc := range []struct {
		name, path, query, want, unwanted string
	}{
		{name: "sign-in", path: "/callback", query: "code=valid", want: "Return to your terminal to finish signing in.", unwanted: "setting up your project"},
		{name: "installation", path: "/", query: "installation_id=123", want: "Return to your terminal to finish setting up your project.", unwanted: "signing in"},
		{name: "failure", path: "/callback", query: "error=%3Cscript%3E", want: "Sign-in was not completed", unwanted: "<script>"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ln, _, err := CallbackListener()
			if err != nil {
				t.Fatal(err)
			}
			defer ln.Close()
			result := make(chan error, 1)
			go func() {
				_, err := waitForCallback(ln, tc.path, func(v url.Values) (string, string) {
					if v.Get("error") != "" {
						return "Sign-in was not completed", "Check your terminal for details."
					}
					if v.Get("installation_id") != "" {
						return "Almost there!", "Return to your terminal to finish setting up your project."
					}
					return "Almost there!", "Return to your terminal to finish signing in."
				}, 2*time.Second, time.Second)
				result <- err
			}()
			resp, err := http.Get("http://" + ln.Addr().String() + tc.path + "?" + tc.query)
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(body), tc.want) || strings.Contains(string(body), tc.unwanted) {
				t.Errorf("response = %q, want %q and no %q", body, tc.want, tc.unwanted)
			}
			if err := <-result; err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestWaitForCallbackTimeout(t *testing.T) {
	ln, _, err := CallbackListener()
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	_, err = waitForCallback(ln, "/", func(url.Values) (string, string) { return "ok", "" }, 50*time.Millisecond, 10*time.Millisecond)
	if err == nil {
		t.Fatal("waitForCallback() = nil error, want timeout")
	}
}
