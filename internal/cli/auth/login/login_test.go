package login

import (
	"strings"
	"testing"
)

func TestGenerateChallenge(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	want := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"

	got := generateChallenge(verifier)
	if got != want {
		t.Errorf("generateChallenge(%q) = %q, want %q", verifier, got, want)
	}
}

func TestGenerateVerifier_Format(t *testing.T) {
	verifier, err := generateVerifier()
	if err != nil {
		t.Fatalf("generateVerifier() = %v, want nil", err)
	}

	if verifier == "" {
		t.Fatal(`generateVerifier() = "", want non-empty string`)
	}
	if len(verifier) != 43 {
		t.Errorf("len(generateVerifier()) = %d, want 43", len(verifier))
	}

	// Raw URL encoding must not contain '+', '/', or '='.
	for _, ch := range []string{"+", "/", "="} {
		if strings.Contains(verifier, ch) {
			t.Errorf("generateVerifier() = %q contains disallowed character %q", verifier, ch)
		}
	}
}
