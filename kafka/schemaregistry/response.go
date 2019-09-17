package schemaregistry

// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#config
type ResConfig struct {
	// https://docs.confluent.io/current/schema-registry/avro.html#compatibility-types
	Compatibility string `json:"compatibilityLevel"`
}

// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#post--subjects-(string-%20subject)
type ResCheckSubject struct {
	Subject string `json:"subject"` // Name of the subject that this schema is registered under
	ID      int    `json:"id"`      // Globally unique identifier of the schema
	Version int    `json:"version"` // Version of the returned schema
	Schema  string `json:"schema"`  // The Avro schema string
}

// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#post--subjects-(string-%20subject)-versions
type ResRegisterNewSchema struct {
	ID int `json:"id"` // Globally unique identifier of the schema
}

// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#get--subjects-(string-%20subject)-versions
type ResGetSubjectVersionsList []int

// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#get--subjects-(string-%20subject)-versions-(versionId-%20version)
type ResGetSubjectVersion struct {
	Name    string `json:"name"`    // Name of the subject that this schema is registered under
	Version int    `json:"version"` // Version of the returned schema
	Schema  string `json:"schema"`  // The Avro schema string
}

// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#get--subjects
type ResGetSubjectList []string

// https://docs.confluent.io/current/schema-registry/schema-deletion-guidelines.html#schema-deletion-guidelines
type ResDeleteSubject []int

// https://docs.confluent.io/current/schema-registry/schema-deletion-guidelines.html#schema-deletion-guidelines
type ResDeleteSubjectVersion int

// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#get--schemas-ids-int-%20id
type ResSchema struct {
	Schema string `json:"schema"`
}
