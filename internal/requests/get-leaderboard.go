package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/httpclient"
)

type LeaderboardTarget string

const (
	Project      LeaderboardTarget = "project"
	Organization LeaderboardTarget = "org"
)

func GetLeaderboard(endpoint, accessToken, id, to, from string, target LeaderboardTarget) ([]domain.LeaderboardEntry, error) {
	id, err := canonicalID(id)
	if err != nil {
		return []domain.LeaderboardEntry{}, err
	}

	query := url.Values{"to": {to}, "from": {from}}
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/leaderboard/%s/%s?%s", endpoint, target, id, query.Encode()),
		nil,
	)
	if err != nil {
		return []domain.LeaderboardEntry{}, fmt.Errorf("failed to initialize request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := httpclient.Client.Do(req)
	if err != nil {
		return []domain.LeaderboardEntry{}, fmt.Errorf("failed to get leaderboard: %w", httpclient.ClassifyHTTPError(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return []domain.LeaderboardEntry{}, httpclient.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []domain.LeaderboardEntry{}, fmt.Errorf("failed to read get projects response: %w", httpclient.ClassifyHTTPError(err))
	}

	if resp.StatusCode > 299 {
		return []domain.LeaderboardEntry{}, fmt.Errorf("failed to get leaderboard. received response status of '%d': %s", resp.StatusCode, string(body))
	}

	var leaderboard domain.GetLeaderboardResponse
	if err := json.Unmarshal(body, &leaderboard); err != nil {
		return []domain.LeaderboardEntry{}, fmt.Errorf("failed to unmarshal projects from get request: %w", err)
	}

	return leaderboard.Entries, nil
}
