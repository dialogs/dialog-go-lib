package serde

import "context"

type IGetter interface {
	Get(context.Context, Kind, int) (*Deserializer, error)
}
