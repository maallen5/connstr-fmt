package main

import (
	"fmt"
	"strings"
)

// Normalize detects which of the two DSN styles the input uses and
// rewrites it into a canonical form: trimmed whitespace, consistent
// casing, and deterministic ordering of parameters. Two DSNs that
// mean the same thing should normalize to the same string.
func Normalize(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", fmt.Errorf("empty input")
	}

	if strings.Contains(s, "://") {
		return normalizeURL(s)
	}
	if strings.Contains(s, "=") {
		return normalizeKeyValue(s)
	}
	return "", fmt.Errorf("unrecognized connection string format")
}
