package parser

// ValidSchemaVersions available schema versions
var ValidSchemaVersions = []string{
	"0.2",
}

// isValidSchemaVersion validates schema version
func isValidSchemaVersion(schemaVersion string) bool {
	for _, validVersion := range ValidSchemaVersions {
		if schemaVersion == validVersion {
			return true
		}
	}
	return false
}
