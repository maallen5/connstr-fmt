package main

import "testing"

func TestNormalizeURL(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{
			name: "lowercases scheme and host, leaves userinfo alone",
			in:   "postgres://User:pw@DB.Example.COM:5432/app?sslmode=require",
			want: "postgres://User:pw@db.example.com:5432/app?sslmode=require",
		},
		{
			name: "sorts query params alphabetically",
			in:   "postgres://user:pw@host:5432/app?sslmode=require&application_name=svc",
			want: "postgres://user:pw@host:5432/app?application_name=svc&sslmode=require",
		},
		{
			name: "drops trailing dot on hostname",
			in:   "mysql://user:pw@host.example.com.:3306/db",
			want: "mysql://user:pw@host.example.com:3306/db",
		},
		{
			name: "strips trailing slash from path",
			in:   "postgres://host/mydb/",
			want: "postgres://host/mydb",
		},
		{
			name: "root path is left as a single slash",
			in:   "postgres://host/",
			want: "postgres://host/",
		},
		{
			name: "no path is left empty",
			in:   "postgres://host",
			want: "postgres://host",
		},
		{
			name: "already canonical is left unchanged",
			in:   "postgres://user:pw@host:5432/app?application_name=svc&sslmode=require",
			want: "postgres://user:pw@host:5432/app?application_name=svc&sslmode=require",
		},
		{
			name:    "invalid URL is an error",
			in:      "postgres://user:pw@[::1",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Normalize(tc.in, false)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Normalize(%q) = %q, want error", tc.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Normalize(%q) returned error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("Normalize(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeJDBCURL(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{
			name: "lowercases prefix and delegates to URL normalization",
			in:   "JDBC:mysql://User:pw@DB.Example.COM:3306/app?serverTimezone=UTC&useSSL=false",
			want: "jdbc:mysql://User:pw@db.example.com:3306/app?serverTimezone=UTC&useSSL=false",
		},
		{
			name: "already canonical is left unchanged",
			in:   "jdbc:postgresql://host:5432/app?sslmode=require",
			want: "jdbc:postgresql://host:5432/app?sslmode=require",
		},
		{
			name:    "invalid URL after the prefix is an error",
			in:      "jdbc:mysql://user:pw@[::1",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Normalize(tc.in, false)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Normalize(%q) = %q, want error", tc.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Normalize(%q) returned error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("Normalize(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeMySQLDSN(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{
			name: "lowercases protocol and host, sorts params",
			in:   "user:pass@tcp(HOST.Example.COM:3306)/mydb?loc=Local&parseTime=true",
			want: "user:pass@tcp(host.example.com:3306)/mydb?loc=Local&parseTime=true",
		},
		{
			name: "no protocol or address, just a database name",
			in:   "/mydb",
			want: "/mydb",
		},
		{
			name: "no user, plain tcp address",
			in:   "tcp(localhost:3306)/mydb",
			want: "tcp(localhost:3306)/mydb",
		},
		{
			name: "unix socket address keeps its path case",
			in:   "user@unix(/var/run/MySQL/mysqld.sock)/mydb",
			want: "user@unix(/var/run/MySQL/mysqld.sock)/mydb",
		},
		{
			name: "user with no password",
			in:   "user@tcp(host:3306)/mydb",
			want: "user@tcp(host:3306)/mydb",
		},
		{
			name: "already canonical is left unchanged",
			in:   "user:pass@tcp(host:3306)/mydb?loc=Local&parseTime=true",
			want: "user:pass@tcp(host:3306)/mydb?loc=Local&parseTime=true",
		},
		{
			name:    "missing closing paren is an error",
			in:      "tcp(host:3306/mydb",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Normalize(tc.in, false)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Normalize(%q) = %q, want error", tc.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Normalize(%q) returned error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("Normalize(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeKeyValue(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{
			name: "sorts keys and lowercases them",
			in:   "Server = LOCALHOST ; Uid=admin; Database=Orders;pwd=hunter2",
			want: "database=Orders;password=hunter2;server=localhost;user id=admin",
		},
		{
			name: "collapses key aliases to canonical name",
			in:   "Data Source=host;Initial Catalog=db;User=admin;Password=x",
			want: "database=db;password=x;server=host;user id=admin",
		},
		{
			name: "quotes values containing a separator",
			in:   "Server=host;Database=my db",
			want: "database='my db';server=host",
		},
		{
			name: "strips surrounding quotes from input values",
			in:   `Server=host;Database="my db"`,
			want: "database='my db';server=host",
		},
		{
			name: "drops trailing separator and blank segments",
			in:   "Server=host;;Database=db;",
			want: "database=db;server=host",
		},
		{
			name: "last value for a repeated key wins",
			in:   "Server=first;Server=second",
			want: "server=second",
		},
		{
			name:    "empty input is an error",
			in:      "   ",
			wantErr: true,
		},
		{
			name:    "input with neither '://' nor '=' is an error",
			in:      "not a dsn at all",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Normalize(tc.in, false)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Normalize(%q) = %q, want error", tc.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Normalize(%q) returned error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("Normalize(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeRedact(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "URL password is masked, username untouched",
			in:   "postgres://user:hunter2@host:5432/app",
			want: "postgres://user:REDACTED@host:5432/app",
		},
		{
			name: "URL with no password is left alone",
			in:   "postgres://user@host:5432/app",
			want: "postgres://user@host:5432/app",
		},
		{
			name: "URL with no userinfo is left alone",
			in:   "postgres://host:5432/app",
			want: "postgres://host:5432/app",
		},
		{
			name: "JDBC URL password is masked",
			in:   "jdbc:mysql://user:hunter2@host:3306/app",
			want: "jdbc:mysql://user:REDACTED@host:3306/app",
		},
		{
			name: "MySQL driver DSN password is masked",
			in:   "user:hunter2@tcp(host:3306)/mydb",
			want: "user:REDACTED@tcp(host:3306)/mydb",
		},
		{
			name: "MySQL driver DSN with no password is left alone",
			in:   "user@tcp(host:3306)/mydb",
			want: "user@tcp(host:3306)/mydb",
		},
		{
			name: "key=value password is masked",
			in:   "Server=host;Uid=admin;Pwd=hunter2",
			want: "password=REDACTED;server=host;user id=admin",
		},
		{
			name: "key=value with no password is left alone",
			in:   "Server=host;Uid=admin",
			want: "server=host;user id=admin",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Normalize(tc.in, true)
			if err != nil {
				t.Fatalf("Normalize(%q) returned error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("Normalize(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
