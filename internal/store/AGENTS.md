# internal/store — agent notes

Rules for changing this package. Read PROJECT.md for the rationale.

- **Two tables only: `memories` and `calendar_events`.** Do not add tables, and do not add typed columns to `memories`. New kinds of data become rows in `memories` with a new `source` prefix. Calendar events are the single structured exception because the brief needs real datetime range queries on them.
- **`source` is hierarchical: `prefix:instance`.** Examples in use: `weather-metno:Helsinki`, `calendar:Family Calendar`, `schoollunch:<school>`, `electricity:<zone>`, `manual`. The prefix is owned by the importer package that writes it, as an exported `SourcePrefix` constant. The brief filters on these prefixes, so renaming one is a cross-package change.
- **Deduplication is `(source, uid, relevance_date)`** for memories and `(source, uid, start_time)` for calendar events. Importers must supply a stable `uid` if they want upsert behaviour. Nil `uid` means always insert.
- **Schema changes go in `Initialize()` as `CREATE ... IF NOT EXISTS`.** There is no migration system. Adding a column means adding an idempotent `ALTER TABLE` guarded by a check, not editing the `CREATE TABLE`.
- **Pure Go SQLite via `modernc.org/sqlite`.** Never import `mattn/go-sqlite3` or anything needing cgo; the Linux build must stay static.
- **Tests use a real SQLite file in `t.TempDir()`**, see `store_test.go`. That is deterministic and allowed. No shared fixture database.
