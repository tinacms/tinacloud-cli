package login

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/browser"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
)

func generateVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func callbackPort() (int, error) {
	port := 0
	for i := 9999; i <= 10011; i++ {
		timeout := time.Second
		conn, _ := net.DialTimeout("tcp", net.JoinHostPort("localhost", strconv.Itoa(i)), timeout)
		if conn != nil {
			defer conn.Close()
			continue
		}
		port = i
		break
	}
	if port == 0 {
		return port, fmt.Errorf("failed to bind local callback server on port range :9999-:10011")
	}

	return port, nil
}

func Login(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("login", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli auth login")
		fmt.Fprintln(os.Stderr, "\nLog in to a TinaCloud account")
	}
	_ = flagSet.Parse(args)

	verifier, err := generateVerifier()
	if err != nil {
		return fmt.Errorf("failed to generate verifier code: %w", err)
	}
	challenge := generateChallenge(verifier)

	port, err := callbackPort()
	if err != nil {
		return err
	}

	{
		// <tinacloud-host>:oauth/authorize will attempt to redirect the client
		// to the authorization page. instead, we want to catch the redirect
		// location to display to user.
		client := http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		oauthURL := fmt.Sprintf("%s/oauth/authorize?code_challenge=%s&redirect_uri=http://localhost:%d/callback", cfg.TinaCloud.Endpoint, challenge, port)
		resp, err := client.Get(oauthURL)
		if err != nil {
			return fmt.Errorf("failed to authorize with tinacloud: %w", err)
		}

		redirectURL := resp.Header.Get("location")
		if redirectURL == "" {
			// Default to pre-redirect URL if no location header is found
			redirectURL = oauthURL
		}

		// User can copy string from terminal if this fails
		_ = browser.OpenURL(redirectURL)

		fmt.Printf(`TinaCloud SSO login required.
Attempting to open your default browser. If the browser does not open, open the following URL.

%s

`, redirectURL)
	}

	// Read auth code from client
	code := ""
	{
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return fmt.Errorf("failed to start callback server: %w", err)
		}
		defer ln.Close()

		conn, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept auth callback: %w", err)
		}
		defer conn.Close()

		req, err := http.ReadRequest(bufio.NewReader(conn))
		if err != nil {
			return fmt.Errorf("failed to read auth callback: %w", err)
		}

		// Send message back to client
		writer := bufio.NewWriter(conn)
		_, err = writer.WriteString("HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/plain\r\n" +
			"Connection: close\r\n" +
			"\r\n" +
			"You may now close this tab.")
		if err != nil {
			return fmt.Errorf("failed to write back to connection: %w", err)
		}
		err = writer.Flush()
		if err != nil {
			return fmt.Errorf("failed to flush writer to connection: %w", err)
		}

		params, err := url.ParseQuery(req.URL.RawQuery)
		if err != nil {
			return fmt.Errorf("failed to read code from auth callback: %w", err)
		}

		code = params.Get("code")
		if code == "" {
			return errors.New("code is missing from client request")
		}
	}

	// Verify auth code
	var tokens domain.AuthResponse
	{
		params := url.Values{}
		params.Set("grant_type", "authorization_code")
		params.Set("code", code)
		params.Set("redirect_uri", fmt.Sprintf("http://localhost:%d/callback", port))
		params.Set("code_verifier", verifier)

		resp, err := http.PostForm(fmt.Sprintf("%s/oauth/token", cfg.TinaCloud.Endpoint), params)
		if err != nil {
			return fmt.Errorf("failed to post code to cognito for verification: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("failed to authorize user")
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read auth validation response: %w", err)
		}

		if resp.StatusCode > 299 {
			return fmt.Errorf("failed to verify auth token. receieved response status of '%d': %s", resp.StatusCode, string(body))
		}

		if err := json.Unmarshal(body, &tokens); err != nil {
			return fmt.Errorf("could not parse tokens: %w", err)
		}
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		return fmt.Errorf("received empty auth tokens from login request")
	}

	err = auth.SaveTokens(auth.TokenCache{
		IDToken:      tokens.IDToken,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second),
	})
	if err != nil {
		return fmt.Errorf("failed to save tokens to file: %w", err)
	}

	fmt.Println("Successfully logged into your TinaCloud account!")

	return nil
}
