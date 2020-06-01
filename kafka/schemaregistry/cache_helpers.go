package schemaregistry

import (
	"context"

	"github.com/dialogs/dialog-go-lib/serde/avro"
	"github.com/pkg/errors"
)

func NewGCacheSchemaLoaderFunc(c *Client) func(interface{}) (interface{}, error) {
	return func(key interface{}) (interface{}, error) {

		schemaID, ok := key.(int)
		if !ok {
			return nil, errors.Errorf("invalid key type: %T", key)
		}

		schemaDesc, err := c.GetSchema(context.Background(), schemaID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get schema")
		}

		return schemaDesc.Schema, nil
	}
}

func NewGCacheDeserializerLoaderFunc(c *Client) func(interface{}) (interface{}, error) {

	schemaLoader := NewGCacheSchemaLoaderFunc(c)

	return func(key interface{}) (interface{}, error) {

		schema, err := schemaLoader(key)
		if err != nil {
			return nil, err
		}

		schemaVal, ok := schema.(string)
		if !ok {
			return nil, errors.Errorf("invalid schema type: %T", schema)
		}

		des, err := avro.NewDeserializer(schemaVal, schemaVal)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get schema")
		}

		return des, nil
	}
}
