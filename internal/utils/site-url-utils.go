package utils

import (
	"fmt"
	"net/url"
	"strings"
)

// CleanSiteURL normalizes a site URL to its origin (scheme://host[:port]); globs pass through, invalid http(s) inputs error.
func CleanSiteURL(input string) (string, error) {
	input = strings.TrimSpace(input)

	// Allow glob patterns.
	if strings.Contains(input, "*") {
		return input, nil
	}

	parsed, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("invalid URL %q: %w", input, err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("URL %q must use http or https", input)
	}

	if parsed.Host == "" {
		return "", fmt.Errorf("URL %q must include a host", input)
	}

	// Keep only the origin: scheme://host[:port]
	return strings.TrimRight(parsed.Scheme+"://"+parsed.Host, "/"), nil
}

func CleanSiteURLs(inputs []string) ([]string, error) {
	var cleaned []string

	for _, input := range inputs {
		if strings.TrimSpace(input) == "" {
			continue
		}

		origin, err := CleanSiteURL(input)
		if err != nil {
			return nil, err
		}

		cleaned = append(cleaned, origin)
	}

	return cleaned, nil
}
