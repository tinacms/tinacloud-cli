package domain

type GetV2PostHogTokenResponse struct {
	APIKey   string `json:"api_key"`
	Endpoint string `json:"endpoint"`
}
