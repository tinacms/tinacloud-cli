package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/domain"
)

func CreateProject(endpoint, accessToken, orgID string, body domain.PostV2ProjectsBody) (domain.PostV2ProjectResponse, error) {
	postBody, err := json.Marshal(body)
	if err != nil {
		return domain.PostV2ProjectResponse{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	client := http.Client{}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/v2/organizations/%s/apps", endpoint, orgID), bytes.NewBuffer(postBody))
	if err != nil {
		return domain.PostV2ProjectResponse{}, fmt.Errorf("failed to initialize request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return domain.PostV2ProjectResponse{}, fmt.Errorf("failed to create project: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.PostV2ProjectResponse{}, auth.ErrUnauthorized
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.PostV2ProjectResponse{}, fmt.Errorf("failed to read create project response: %w", err)
	}

	if resp.StatusCode > 299 {
		return domain.PostV2ProjectResponse{}, fmt.Errorf("failed to create project. received response status of '%d': %s", resp.StatusCode, string(responseBody))
	}

	var projectCreateResponse domain.PostV2ProjectResponse
	if err := json.Unmarshal(responseBody, &projectCreateResponse); err != nil {
		return domain.PostV2ProjectResponse{}, fmt.Errorf("failed to unmarshal created project from post request: %w", err)
	}

	return projectCreateResponse, nil
}
