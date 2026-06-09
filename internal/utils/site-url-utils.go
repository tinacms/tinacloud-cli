package utils

import (
	"fmt"
	"net/url"
	"strings"
)

func CleanSiteURLs(inputs []string) ([]string, error) {
	var cleaned []string

	for _, input := range inputs {
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Allow glob patterns.
		if strings.Contains(input, "*") {
			cleaned = append(cleaned, input)
			continue
		}

		parsed, err := url.Parse(input)
		if err != nil {
			return nil, fmt.Errorf("invalid URL %q: %w", input, err)
		}

		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return nil, fmt.Errorf("URL %q must use http or https", input)
		}

		if parsed.Host == "" {
			return nil, fmt.Errorf("URL %q must include a host", input)
		}

		// Keep only the origin: scheme://host[:port]
		origin := parsed.Scheme + "://" + parsed.Host
		origin = strings.TrimRight(origin, "/")

		cleaned = append(cleaned, origin)
	}

	return cleaned, nil
}
