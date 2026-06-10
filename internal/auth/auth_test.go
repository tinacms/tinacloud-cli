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
		name       string
		token      string
		wantUserID string
		wantErr    bool
	}{
		{
			name: "valid token with tinacloud_user_id",
			token: makeJWT(t, &DecodedAccessToken{
				TinaCloudUserID:  "user-abc",
				RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1"},
			}),
			wantUserID: "user-abc",
		},
		{
			name: "valid token without tinacloud_user_id",
			token: makeJWT(t, &DecodedAccessToken{
				RegisteredClaims: jwt.RegisteredClaims{Subject: "user-2"},
			}),
			wantUserID: "",
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
			if got.TinaCloudUserID != tc.wantUserID {
				t.Errorf("DecodeAccessToken(%q).TinaCloudUserID = %q, want %q", tc.token, got.TinaCloudUserID, tc.wantUserID)
			}
		})
	}
}
