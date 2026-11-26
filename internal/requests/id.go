package requests

import (
	"fmt"
	"uuid"
)

func canonicalID(id string) (string, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return "", fmt.Errorf("invalid UUID: %w", err)
	}
	return parsed.String(), nil
}
