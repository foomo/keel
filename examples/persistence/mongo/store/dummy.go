package store

type Dummy struct {
	Entity               `json:",inline" bson:",inline"`
	EntityWithVersions   `json:",inline" bson:",inline"`
	EntityWithTimestamps `json:",inline" bson:",inline"`
}
