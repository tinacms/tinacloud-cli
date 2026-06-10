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

// TestDecodeAccessToken_WireClaimName locks the on-the-wire JSON claim name.
//
// The table-driven test above mints tokens through the DecodedAccessToken
// struct, so both encode and decode use the same json tag — it would still
// pass if the tag were silently reverted to "tinacloud_client_id". This test
// mints from a raw jwt.MapClaims with the literal claim string the identity
// service emits (see tinacms/tinacloud#3677), so it fails if the struct tag
// drifts away from what the producer actually sends.
func TestDecodeAccessToken_WireClaimName(t *testing.T) {
	mint := func(t *testing.T, claims jwt.MapClaims) string {
		t.Helper()
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := tok.SignedString([]byte("test-secret"))
		if err != nil {
			t.Fatalf("mint: %v", err)
		}
		return signed
	}

	t.Run("decodes the tinacloud_user_id wire claim", func(t *testing.T) {
		token := mint(t, jwt.MapClaims{"tinacloud_user_id": "user-abc"})
		got, err := DecodeAccessToken(token)
		if err != nil {
			t.Fatalf("DecodeAccessToken() = %v, want nil error", err)
		}
		if got.TinaCloudUserID != "user-abc" {
			t.Errorf("TinaCloudUserID = %q, want %q", got.TinaCloudUserID, "user-abc")
		}
	})

	t.Run("ignores the legacy tinacloud_client_id wire claim", func(t *testing.T) {
		// Regression guard: a token carrying only the old claim name must not
		// populate TinaCloudUserID. If this fails, the struct tag was reverted.
		token := mint(t, jwt.MapClaims{"tinacloud_client_id": "user-abc"})
		got, err := DecodeAccessToken(token)
		if err != nil {
			t.Fatalf("DecodeAccessToken() = %v, want nil error", err)
		}
		if got.TinaCloudUserID != "" {
			t.Errorf("TinaCloudUserID = %q, want empty (old claim must not bind)", got.TinaCloudUserID)
		}
	})
}
