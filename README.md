# connstr-fmt

Connection strings collect entropy over time. People paste them between
config files, tweak casing, add stray whitespace, reorder query
parameters, and copy the ODBC form into a place that expects the URL
form. Two connection strings that mean exactly the same thing end up
looking different, which makes them annoying to diff, dedupe, or grep
for in logs and config repos.

`connfmt` takes a messy connection string and rewrites it into a
canonical form: consistent casing, trimmed whitespace, and
deterministically ordered parameters. It does not validate that a
connection string will actually connect to anything - it just makes
equivalent strings compare equal as text.

It handles the two DSN shapes you run into most:

- URL style: `postgres://user:pass@Host:5432/mydb?sslmode=require&connect_timeout=10`
- ADO.NET / ODBC key=value style: `Server=host;Database=db;User Id=admin;Password=x`

## Usage

Build it:

```
go build -o connfmt .
```

Pass a connection string as an argument:

```
$ ./connfmt "Server = LOCALHOST ; Uid=admin; Database=Orders;pwd=hunter2"
database=Orders;password=hunter2;server=localhost;user id=admin
```

Or pipe it in on stdin:

```
$ echo 'postgres://User:pw@DB.Example.COM:5432/app?sslmode=require&application_name=svc' | ./connfmt
postgres://User:pw@db.example.com:5432/app?application_name=svc&sslmode=require
```

What it currently normalizes:

- scheme and hostname are lowercased (userinfo and path segments are
  left alone, since those are often case-sensitive)
- surrounding whitespace around keys and values is stripped
- query parameters and key=value pairs are sorted alphabetically
- known key aliases (`uid`, `user`, `username` -> `user id`; `pwd` ->
  `password`; `data source` -> `server`; `initial catalog` ->
  `database`) collapse to one canonical name
- values are quoted only when they contain a separator character

## Status

Early. The parser handles the common cases above but doesn't yet know
about every driver-specific quirk. See the roadmap in the commit
history for what's planned next.

## License

MIT, see [LICENSE](LICENSE).
