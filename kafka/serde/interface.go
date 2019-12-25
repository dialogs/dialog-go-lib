package serde

import (
	"context"
)

type IGetter interface {
	Get(ctx context.Context, kind Kind, schemaID int) (*Deserializer, error)
}
