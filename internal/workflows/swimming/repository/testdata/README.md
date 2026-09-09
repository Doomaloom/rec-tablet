# Python compatibility fixture

These synthetic records were produced by the original Python `Store` before its removal, using `examples/classes.json`.

- `python-v1.sql`: Python sqlite3 `iterdump()` output plus the original `PRAGMA user_version = 1` (the dump API omits this pragma).
- `python-v1-export.json`: the original exporter snapshot, including its JSON string columns.
- `identities.json`: device-generated IDs needed to continue the original records in tests.

The fixture includes attendance, three originally successful attempts, a submitted correction and its revision, an unfinished correction, and an unfinished five-criterion assessment. `TestPythonDatabaseCompatibility` restores the SQLite dump, opens it through the Go store without importing a class package, compares every exported table, and edits the original unfinished work. The tests require no Python installation.
