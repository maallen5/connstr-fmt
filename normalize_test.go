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
			got, err := Normalize(tc.in)
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
			got, err := Normalize(tc.in)
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
