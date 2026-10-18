package transfer

// Snapshot is the existing v1 table export contract, independent of domain models.
type Snapshot struct {
	SchemaVersion   int                         `json:"schemaVersion"`
	Kind            string                      `json:"kind"`
	ExportedAt      string                      `json:"exportedAt"`
	DatabaseVersion int                         `json:"databaseVersion"`
	Tables          map[string][]map[string]any `json:"tables"`
}
