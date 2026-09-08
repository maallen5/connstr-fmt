package main

import (
	"fmt"
	"strings"
)

// looksLikeMySQLDSN reports whether s matches the DSN format the
// MySQL driver itself expects, e.g.
//
//	user:pass@tcp(localhost:3306)/dbname?parseTime=true&loc=Local
//
// That's neither a URL (no "://") nor an ADO.NET key=value string
// (nothing before the first "/" looks like "key="), so it needs its
// own detection ahead of the other two.
func looksLikeMySQLDSN(s string) bool {
	if strings.Contains(s, "://") || strings.Contains(s, ";") {
		return false
	}
	i := strings.Index(s, "/")
	if i < 0 {
		return false
	}
	return !strings.Contains(s[:i], "=")
}

type mysqlDSN struct {
	user, pass, net, addr, dbname, params string
}

// parseMySQLDSN splits a MySQL driver DSN into its parts:
//
//	[user[:pass]@][net[(addr)]]/dbname[?param1=value1&...]
//
// It finds the database name by looking for the last "/" in the
// string, which works even when addr is a unix socket path that
// contains its own "/" (dbname itself never does).
func parseMySQLDSN(s string) (mysqlDSN, error) {
	var d mysqlDSN

	slash := strings.LastIndexByte(s, '/')
	if slash < 0 {
		return d, fmt.Errorf("no '/' separating address from database name")
	}

	head := s[:slash]
	tail := s[slash+1:]

	if q := strings.IndexByte(tail, '?'); q >= 0 {
		d.dbname = tail[:q]
		d.params = tail[q+1:]
	} else {
		d.dbname = tail
	}

	if at := strings.LastIndexByte(head, '@'); at >= 0 {
		userinfo := head[:at]
		head = head[at+1:]
		if c := strings.IndexByte(userinfo, ':'); c >= 0 {
			d.user = userinfo[:c]
			d.pass = userinfo[c+1:]
		} else {
			d.user = userinfo
		}
	}

	if p := strings.IndexByte(head, '('); p >= 0 {
		if !strings.HasSuffix(head, ")") {
			return d, fmt.Errorf("address missing closing ')'")
		}
		d.net = head[:p]
		d.addr = head[p+1 : len(head)-1]
	} else {
		d.net = head
	}

	return d, nil
}

// normalizeMySQLDSN rewrites a MySQL driver DSN into canonical form:
// lowercased network protocol and hostname, sorted query parameters.
func normalizeMySQLDSN(s string) (string, error) {
	d, err := parseMySQLDSN(s)
	if err != nil {
		return "", err
	}

	d.net = strings.ToLower(d.net)
	d.addr = normalizeMySQLAddr(d.addr)

	var b strings.Builder
	if d.user != "" || d.pass != "" {
		b.WriteString(d.user)
		if d.pass != "" {
			b.WriteByte(':')
			b.WriteString(d.pass)
		}
		b.WriteByte('@')
	}
	b.WriteString(d.net)
	if d.addr != "" {
		b.WriteByte('(')
		b.WriteString(d.addr)
		b.WriteByte(')')
	}
	b.WriteByte('/')
	b.WriteString(d.dbname)

	if d.params != "" {
		b.WriteByte('?')
		b.WriteString(sortedQuery(d.params))
	}

	return b.String(), nil
}

// normalizeMySQLAddr lowercases the hostname of a tcp address while
// leaving a unix socket path untouched, since paths are case-sensitive
// on most filesystems and hostnames aren't.
func normalizeMySQLAddr(addr string) string {
	if addr == "" || strings.HasPrefix(addr, "/") {
		return addr
	}
	host := addr
	port := ""
	if i := strings.LastIndex(addr, ":"); i >= 0 {
		host = addr[:i]
		port = addr[i:]
	}
	return strings.ToLower(host) + port
}
