package store

// Dummy is an example entity with an ID, a version and timestamps.
type Dummy struct {
	Entity               `json:",inline" bson:",inline"`
	EntityWithVersions   `json:",inline" bson:",inline"`
	EntityWithTimestamps `json:",inline" bson:",inline"`
}
