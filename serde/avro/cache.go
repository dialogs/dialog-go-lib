package avro

import (
	"context"
	"sync"

	"github.com/dialogs/dialog-go-lib/kafka/schemaregistry"
	"github.com/pkg/errors"
)

type Cache struct {
	schemas        map[Kind]map[int]*Deserializer
	schemaregistry *schemaregistry.Client
	mu             sync.RWMutex
}

func NewCache(client *schemaregistry.Client) *Cache {
	return &Cache{
		schemas:        make(map[Kind]map[int]*Deserializer),
		schemaregistry: client,
	}
}

func (d *Cache) Get(ctx context.Context, kind Kind, schemaID int) (*Deserializer, error) {

	schemas := d.getSchemas(kind)

	d.mu.RLock()
	des, ok := schemas[schemaID]
	d.mu.RUnlock()
	if ok {
		return des, nil
	}

	schema, err := d.schemaregistry.GetSchema(ctx, schemaID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get schema for '%s'", kind.String())
	}

	des, err = New(schema.Schema, schema.Schema)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create deserializer for '%s'", kind.String())
	}

	d.mu.Lock()
	schemas[schemaID] = des
	d.mu.Unlock()

	return des, nil
}

func (d *Cache) getSchemas(kind Kind) map[int]*Deserializer {

	d.mu.Lock()
	defer d.mu.Unlock()

	schemas, ok := d.schemas[kind]
	if ok {
		return schemas
	}

	schemas = make(map[int]*Deserializer)
	d.schemas[kind] = schemas

	return schemas
}
