# knbn

No bullshit 1-file kanban boards.

## Developing

You'll need the following tools:

- [Go](https://go.dev)
- [templ](https://templ.guide)

If you modify any of the `*.templ` files in `templs` then you need to generate
the new `*.go` files for them before compiling.
```
> templ generate
Processing path: /usr/local/src/limeleaf-coop/knbn
Generating production code: /usr/local/src/limeleaf-coop/knbn
(✓) Generated code for "/usr/local/src/limeleaf-coop/knbn/templs/boards.templ" in 2.674596ms
(✓) Generated code for "/usr/local/src/limeleaf-coop/knbn/templs/layout.templ" in 4.892115ms
(✓) Generated code for 2 templates with 0 errors in 5.085521ms
```

## Running

```
> go run ./cmd/main.go
```

Open http://localhost:8080

## Seeding Database

If you want to seed the database with some data you can run:

```
> go run ./cmd/main.go -seed-data-dir ./testdata/db
```

Feel free to add more `.json` files for more data. To re-seed a new databse just
delete the database file on disk first or else you'll see key contraint errors.
