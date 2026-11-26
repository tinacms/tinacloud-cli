package requests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"tina.io/tinacloud-cli/internal/domain"
	"tina.io/tinacloud-cli/internal/httpclient"
)

var ErrMissingInstallationID error = errors.New("failed to retrieve installation ID")

func CreateProject(endpoint, accessToken, orgID string, body domain.PostV2ProjectsBody) (domain.PostV2ProjectResponse, error) {
	if err := validateIDs(orgID); err != nil {
		return domain.PostV2ProjectResponse{}, err
	}

	postBody, err := json.Marshal(body)
	if err != nil {
		return domain.PostV2ProjectResponse{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/v2/organizations/%s/apps", endpoint, orgID), bytes.NewBuffer(postBody))
	if err != nil {
		return domain.PostV2ProjectResponse{}, fmt.Errorf("failed to initialize request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpclient.Client.Do(req)
	if err != nil {
		return domain.PostV2ProjectResponse{}, fmt.Errorf("failed to create project: %w", httpclient.ClassifyHTTPError(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.PostV2ProjectResponse{}, httpclient.ErrUnauthorized
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.PostV2ProjectResponse{}, fmt.Errorf("failed to read create project response: %w", httpclient.ClassifyHTTPError(err))
	}

	if resp.StatusCode > 299 {
		if strings.Contains(string(responseBody), "failed to retrieve installation ID") {
			return domain.PostV2ProjectResponse{}, ErrMissingInstallationID
		}
		return domain.PostV2ProjectResponse{}, fmt.Errorf("failed to create project. received response status of '%d': %s", resp.StatusCode, string(responseBody))
	}

	var projectCreateResponse domain.PostV2ProjectResponse
	if err := json.Unmarshal(responseBody, &projectCreateResponse); err != nil {
		return domain.PostV2ProjectResponse{}, fmt.Errorf("failed to unmarshal created project from post request: %w", err)
	}

	return projectCreateResponse, nil
}
