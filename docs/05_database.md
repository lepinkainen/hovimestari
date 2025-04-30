# Database Schema

The SQLite database (`memories.db`) has a single table called `memories` with the following schema:

```sql
CREATE TABLE memories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    relevance_date TIMESTAMP,
    source TEXT NOT NULL,
    uid TEXT
);
```

Indexes are created on `relevance_date`, `source`, and the combination of `source` and `uid` to optimize queries.
