package utils

import (
	"slices"
	"testing"
)

func TestCleanSiteURLs(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    []string
		wantErr bool
	}{
		{
			name:  "strips path and trailing slash to origin",
			input: []string{"https://example.com/blog/", "http://localhost:3000/"},
			want:  []string{"https://example.com", "http://localhost:3000"},
		},
		{
			name:  "skips blank entries",
			input: []string{"  ", "", "https://example.com"},
			want:  []string{"https://example.com"},
		},
		{
			name:  "passes globs through untouched",
			input: []string{"https://*.example.com"},
			want:  []string{"https://*.example.com"},
		},
		{
			name:    "rejects non-http scheme",
			input:   []string{"ftp://example.com"},
			wantErr: true,
		},
		{
			name:    "rejects missing host",
			input:   []string{"https://"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CleanSiteURLs(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("CleanSiteURLs(%v) = nil error, want error", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("CleanSiteURLs(%v) = %v, want nil error", tc.input, err)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("CleanSiteURLs(%v) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
