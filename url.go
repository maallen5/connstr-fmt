package main

import (
	"net/url"
	"sort"
	"strings"
)

// normalizeURL handles DSNs written as URLs, e.g.
//
//	postgres://user:pass@localhost:5432/mydb?sslmode=require
func normalizeURL(s string, redact bool) (string, error) {
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

	if redact && u.User != nil {
		if _, hasPassword := u.User.Password(); hasPassword {
			u.User = url.UserPassword(u.User.Username(), "REDACTED")
		}
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
func normalizeJDBCURL(s string, redact bool) (string, error) {
	rest := s[len("jdbc:"):]
	normalized, err := normalizeURL(rest, redact)
	if err != nil {
		return "", err
	}
	return "jdbc:" + normalized, nil
}

// normalizeHost lowercases the hostname portion of u.Host and drops a
// trailing dot that some tools add for the "fully qualified" form.
//
// Some drivers (libpq, the Mongo driver, ClickHouse) put a
// comma-separated list of hosts here for replica sets or failover, e.g.
//
//	postgres://user@host1:5432,host2:5433,host3:5434/db
//
// Each entry is normalized on its own and the list order is left as-is,
// since for failover targets the order can affect which host is tried
// first.
//
// Note this relies on net/url having accepted the authority in the
// first place: it only parses a port off the final comma-separated
// entry, so a host list is only valid input here if every entry carries
// a port or none of them do.
func normalizeHost(host string) string {
	if !strings.Contains(host, ",") {
		return normalizeHostPort(host)
	}
	parts := strings.Split(host, ",")
	for i, p := range parts {
		parts[i] = normalizeHostPort(strings.TrimSpace(p))
	}
	return strings.Join(parts, ",")
}

// normalizeHostPort lowercases a single "host" or "host:port" entry.
func normalizeHostPort(host string) string {
	host = strings.TrimSpace(host)
	hostname := host
	port := ""
	if i := strings.LastIndex(host, ":"); i >= 0 {
		hostname = host[:i]
		port = host[i:]
	}
	hostname = strings.ToLower(strings.TrimSuffix(hostname, "."))
	return hostname + port
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
