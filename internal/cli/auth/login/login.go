package login

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/browser"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/httpclient"
	"tina.io/tinacloud-cli/internal/posthog"
	"tina.io/tinacloud-cli/internal/utils"
)

func Command(cfg config.Config, track posthog.Tracker, identify posthog.Identifier) *command.Command {
	return &command.Command{
		Name:        "login",
		Description: "Log in to a TinaCloud account",
		Run: func(args []string) error {
			return runLogin(cfg, browser.OpenURL, track, identify)
		},
	}
}

func runLogin(cfg config.Config, openURL func(string) error, track posthog.Tracker, identify posthog.Identifier) error {
	verifier, err := generateVerifier()
	if err != nil {
		return fmt.Errorf("failed to generate verifier code: %w", err)
	}
	challenge := generateChallenge(verifier)

	ln, port, err := utils.CallbackListener()
	if err != nil {
		return err
	}
	defer ln.Close()

	{
		// <tinacloud-host>:oauth/authorize will attempt to redirect the client
		// to the authorization page. instead, we want to catch the redirect
		// location to display to user.
		client := http.Client{
			Timeout: httpclient.Client.Timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		oauthURL := fmt.Sprintf("%s/oauth/authorize?code_challenge=%s&redirect_uri=http://localhost:%d/callback", cfg.TinaCloud.Endpoint, challenge, port)
		resp, err := client.Get(oauthURL)
		if err != nil {
			return fmt.Errorf("failed to authorize with tinacloud: %w", httpclient.ClassifyHTTPError(err))
		}

		redirectURL := resp.Header.Get("location")
		if redirectURL == "" {
			// Default to pre-redirect URL if no location header is found
			redirectURL = oauthURL
		}

		// User can copy string from terminal if this fails
		_ = openURL(redirectURL)

		track(posthog.UserSignInInitiated())

		fmt.Printf(`Attempting to open your default browser. If the browser does not open, open the following URL.

%s

`, redirectURL)
	}

	// Read auth code from client
	code := ""
	{
		conn, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept auth callback: %w", err)
		}
		defer conn.Close()

		req, err := http.ReadRequest(bufio.NewReader(conn))
		if err != nil {
			return fmt.Errorf("failed to read auth callback: %w", err)
		}

		err = utils.CallbackResponse(conn)
		if err != nil {
			return err
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

		resp, err := httpclient.Client.PostForm(fmt.Sprintf("%s/oauth/token", cfg.TinaCloud.Endpoint), params)
		if err != nil {
			return fmt.Errorf("failed to post code to cognito for verification: %w", httpclient.ClassifyHTTPError(err))
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return httpclient.ErrUnauthorized
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read auth validation response: %w", httpclient.ClassifyHTTPError(err))
		}

		if resp.StatusCode > 299 {
			return fmt.Errorf("failed to verify auth token. received response status of '%d': %s", resp.StatusCode, string(body))
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

	// Attribute the pre-login sign-in events to the now-authenticated user (best-effort).
	if claims, err := auth.DecodeAccessToken(tokens.AccessToken); err == nil {
		identify(claims.TinaCloudUserID)
	}

	track(posthog.UserSignInSuccess())

	fmt.Println("✅️ Successfully logged into TinaCloud!")

	return nil
}

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
