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
			name: "valid token with external_id",
			token: makeJWT(t, &DecodedAccessToken{
				TinaCloudUserID:  "user-abc",
				RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1"},
			}),
			wantUserID: "user-abc",
		},
		{
			name: "valid token without external_id",
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

// TestDecodeAccessToken_WireClaimName pins the literal on-the-wire claim
// string. The table test above mints and decodes through DecodedAccessToken,
// so it round-trips under whatever json tag the struct carries — it can't
// catch a tag drift away from what the identity service emits (#3677).
func TestDecodeAccessToken_WireClaimName(t *testing.T) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"external_id": "user-abc"})
	token, err := tok.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	got, err := DecodeAccessToken(token)
	if err != nil {
		t.Fatalf("DecodeAccessToken() = %v, want nil error", err)
	}
	if got.TinaCloudUserID != "user-abc" {
		t.Errorf("TinaCloudUserID = %q, want %q", got.TinaCloudUserID, "user-abc")
	}
}
