# internal/importer — agent notes

Every data source is a package here. "Add importers, not machinery" is the project direction, so a new capability is almost always a new importer plus a prompt change.

- **Mirror the calendar importer's shape.** One package per source with `type Importer struct{ store *store.Store; ... }`, `func NewImporter(store *store.Store, ...) *Importer`, and `func (i *Importer) Import(ctx context.Context) error`. There is deliberately no shared interface; do not add one until something consumes it.
- **Export a `SourcePrefix` constant** and write memories with `source = fmt.Sprintf("%s:%s", SourcePrefix, instance)`. The brief filters on these prefixes (`internal/brief/brief.go`), so pick the prefix once and keep it.
- **Write to `memories`, not a new table.** Format the data as a human-readable string the LLM can use directly; the content column is what the model sees. Set `relevance_date` to the day the information is about, and a stable `uid` if re-imports should replace rather than duplicate.
- **Calendar is the exception** that writes `calendar_events`. Its `update_mode` config value `smart` maps to upsert and anything else maps to replace-all. gocal's default window is 90 days; the importer widens it, keep that if touching the parser setup.
- **Tests: no network.** Serve fixture payloads from `httptest.NewServer` and point the importer's URL at it; store into a temp SQLite via `internal/store`. Keep the URL or HTTP client injectable for exactly this reason.
- **Wiring:** a new importer also needs a command in `cmd/hovimestari/commands/`, config fields in `internal/config/viper.go` plus `config.example.json`, and usually a prompt instruction in `prompts.json` telling the model what to do with the new memories.
