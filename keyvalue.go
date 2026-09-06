package main

import (
	"fmt"
	"sort"
	"strings"
)

// normalizeKeyValue handles the ADO.NET / ODBC style DSNs built from
// semicolon-separated key=value pairs, e.g.
//
//	Server=localhost; Database = mydb ;User Id=admin;Password=x
func normalizeKeyValue(s string) (string, error) {
	pairs := splitPairs(s)
	if len(pairs) == 0 {
		return "", fmt.Errorf("no key=value pairs found")
	}

	values := make(map[string]string, len(pairs))
	var order []string
	for _, p := range pairs {
		key, val, ok := splitPair(p)
		if !ok {
			continue
		}
		key = normalizeKey(key)
		val = unquote(strings.TrimSpace(val))
		if key == "server" {
			// Hostnames are case-insensitive; lowercase them here so
			// this matches the host normalization the URL style gets.
			val = strings.ToLower(val)
		}
		if _, exists := values[key]; !exists {
			order = append(order, key)
		}
		values[key] = val
	}
	if len(order) == 0 {
		return "", fmt.Errorf("no key=value pairs found")
	}

	sort.Strings(order)

	parts := make([]string, 0, len(order))
	for _, k := range order {
		parts = append(parts, k+"="+quoteIfNeeded(values[k]))
	}
	return strings.Join(parts, ";"), nil
}

// splitPairs breaks a semicolon-delimited DSN into raw "key=value"
// segments, dropping empty segments caused by trailing separators.
func splitPairs(s string) []string {
	raw := strings.Split(s, ";")
	pairs := make([]string, 0, len(raw))
	for _, r := range raw {
		r = strings.TrimSpace(r)
		if r != "" {
			pairs = append(pairs, r)
		}
	}
	return pairs
}

func splitPair(p string) (key, val string, ok bool) {
	i := strings.Index(p, "=")
	if i < 0 {
		return "", "", false
	}
	return p[:i], p[i+1:], true
}

// normalizeKey trims and lowercases the key, then collapses the
// handful of aliases different drivers use for the same setting down
// to one canonical name.
func normalizeKey(k string) string {
	k = strings.ToLower(strings.TrimSpace(k))
	switch k {
	case "server", "data source", "addr", "address", "network address":
		return "server"
	case "database", "initial catalog":
		return "database"
	case "uid", "user id", "user", "username":
		return "user id"
	case "pwd", "password":
		return "password"
	default:
		return k
	}
}

func unquote(v string) string {
	if len(v) >= 2 {
		if (v[0] == '\'' && v[len(v)-1] == '\'') || (v[0] == '"' && v[len(v)-1] == '"') {
			return v[1 : len(v)-1]
		}
	}
	return v
}

// quoteIfNeeded wraps values containing the separators that would
// otherwise make the output ambiguous when read back in.
func quoteIfNeeded(v string) string {
	if strings.ContainsAny(v, ";= ") {
		return "'" + v + "'"
	}
	return v
}
