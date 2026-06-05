package auth

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

type DecodedAccessToken struct {
	TinaCloudClientID string `json:"tinacloud_client_id"`
	jwt.RegisteredClaims
}

func DecodeAccessToken(accessToken string) (DecodedAccessToken, error) {
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(accessToken, &DecodedAccessToken{})
	if err != nil {
		return DecodedAccessToken{}, fmt.Errorf("failed to parse claims: %w", err)
	}

	claims, ok := token.Claims.(*DecodedAccessToken)
	if !ok {
		return DecodedAccessToken{}, errors.New("failed to coerce token claims")
	}

	return *claims, nil
}
