package main

import (
	"fmt"
	"strings"
)

// Normalize detects which DSN style the input uses (URL, ADO.NET
// key=value, MySQL driver, or JDBC) and rewrites it into a canonical
// form: trimmed whitespace, consistent casing, and deterministic
// ordering of parameters. Two DSNs that mean the same thing should
// normalize to the same string.
//
// If redact is true, password values are replaced with "REDACTED" in
// the output instead of being carried through as-is.
func Normalize(raw string, redact bool) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", fmt.Errorf("empty input")
	}

	if strings.HasPrefix(strings.ToLower(s), "jdbc:") {
		return normalizeJDBCURL(s, redact)
	}
	if strings.Contains(s, "://") {
		return normalizeURL(s, redact)
	}
	if looksLikeMySQLDSN(s) {
		return normalizeMySQLDSN(s, redact)
	}
	if strings.Contains(s, "=") {
		return normalizeKeyValue(s, redact)
	}
	return "", fmt.Errorf("unrecognized connection string format")
}
