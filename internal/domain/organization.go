package domain

type Organization struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	BillingEnabled bool   `json:"billingEnabled"`
}
