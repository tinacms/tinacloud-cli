package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v4"
)

func TestDecodeAccessToken(t *testing.T) {
	makeJWT := func(t *testing.T, claims *DecodedAccessToken) string {
		t.Helper()
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := tok.SignedString([]byte("test-secret"))
		if err != nil {
			t.Fatalf("makeJWT: %v", err)
		}
		return signed
	}

	tests := []struct {
		name         string
		token        string
		wantClientID string
		wantErr      bool
	}{
		{
			name: "valid token with tinacloud_client_id",
			token: makeJWT(t, &DecodedAccessToken{
				TinaCloudClientID: "client-abc",
				RegisteredClaims:  jwt.RegisteredClaims{Subject: "user-1"},
			}),
			wantClientID: "client-abc",
		},
		{
			name: "valid token without tinacloud_client_id",
			token: makeJWT(t, &DecodedAccessToken{
				RegisteredClaims: jwt.RegisteredClaims{Subject: "user-2"},
			}),
			wantClientID: "",
		},
		{
			name:    "malformed token",
			token:   "not.a.jwt",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DecodeAccessToken(tc.token)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("DecodeAccessToken(%q) = nil error, want error", tc.token)
				}
				return
			}

			if err != nil {
				t.Fatalf("DecodeAccessToken(%q) = %v, want nil error", tc.token, err)
			}
			if got.TinaCloudClientID != tc.wantClientID {
				t.Errorf("DecodeAccessToken(%q).TinaCloudClientID = %q, want %q", tc.token, got.TinaCloudClientID, tc.wantClientID)
			}
		})
	}
}
