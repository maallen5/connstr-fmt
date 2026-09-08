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
func Normalize(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", fmt.Errorf("empty input")
	}

	if strings.HasPrefix(strings.ToLower(s), "jdbc:") {
		return normalizeJDBCURL(s)
	}
	if strings.Contains(s, "://") {
		return normalizeURL(s)
	}
	if looksLikeMySQLDSN(s) {
		return normalizeMySQLDSN(s)
	}
	if strings.Contains(s, "=") {
		return normalizeKeyValue(s)
	}
	return "", fmt.Errorf("unrecognized connection string format")
}
