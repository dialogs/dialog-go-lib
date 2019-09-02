package registry

// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#config
type ReqConfig struct {
	// https://docs.confluent.io/current/schema-registry/avro.html#compatibility-types
	Compatibility string `json:"compatibility"`
}

// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#post--subjects-(string-%20subject)-versions
type ReqSubject struct {
	Schema string `json:"schema"`
}
