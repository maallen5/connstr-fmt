package main

import (
	"net/url"
	"sort"
	"strings"
)

// normalizeURL handles DSNs written as URLs, e.g.
//
//	postgres://user:pass@localhost:5432/mydb?sslmode=require
func normalizeURL(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}

	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = normalizeHost(u.Host)

	if u.Path != "" && u.Path != "/" {
		u.Path = "/" + strings.Trim(u.Path, "/")
	}

	if u.RawQuery != "" {
		u.RawQuery = sortedQuery(u.RawQuery)
	}

	return u.String(), nil
}

// normalizeJDBCURL handles JDBC-style URLs, which wrap an ordinary
// connection URL behind a "jdbc:" prefix, e.g.
//
//	jdbc:mysql://user:pass@localhost:3306/mydb?useSSL=false
//
// The part after the prefix is normalized the same way as a plain
// URL; the prefix itself is lowercased and reattached.
func normalizeJDBCURL(s string) (string, error) {
	rest := s[len("jdbc:"):]
	normalized, err := normalizeURL(rest)
	if err != nil {
		return "", err
	}
	return "jdbc:" + normalized, nil
}

// normalizeHost lowercases the hostname portion while leaving
// userinfo and port untouched, and drops a trailing dot that some
// tools add for the "fully qualified" form.
func normalizeHost(host string) string {
	host = strings.TrimSpace(host)
	prefix := ""
	if at := strings.LastIndex(host, "@"); at >= 0 {
		prefix = host[:at+1]
		host = host[at+1:]
	}
	hostname := host
	port := ""
	if i := strings.LastIndex(host, ":"); i >= 0 {
		hostname = host[:i]
		port = host[i:]
	}
	hostname = strings.ToLower(strings.TrimSuffix(hostname, "."))
	return prefix + hostname + port
}

// sortedQuery re-encodes the query string with parameters in
// alphabetical order so equivalent DSNs compare equal as plain text.
func sortedQuery(raw string) string {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return raw
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(values))
	for _, k := range keys {
		for _, v := range values[k] {
			parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v))
		}
	}
	return strings.Join(parts, "&")
}
