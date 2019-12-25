package serde

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/actgardner/gogen-avro/compiler"
	"github.com/actgardner/gogen-avro/vm"
	"github.com/actgardner/gogen-avro/vm/types"
	"github.com/pkg/errors"
)

const _SchemaIDSize = 4

type Deserializer struct {
	*vm.Program
}

func New(fromSchema, toSchema string) (*Deserializer, error) {

	p, err := compiler.CompileSchemaBytes([]byte(fromSchema), []byte(toSchema))
	if err != nil {
		return nil, err
	}

	return &Deserializer{
		Program: p,
	}, nil
}

func (d Deserializer) Decode(r io.Reader, field types.Field) error {
	return vm.Eval(r, d.Program, field)
}

func GetDeserializer(ctx context.Context, schemaKind Kind, src []byte, deserializers IGetter) (*Deserializer, []byte, error) {

	if len(src) < _SchemaIDSize+1 { // +1 magic byte
		return nil, nil, fmt.Errorf("invalid incoming avro request: %#v", src)
	}

	if src[0] != 0x00 {
		return nil, nil, fmt.Errorf("invalid magic byte in incoming avro request: %#v", src)
	}

	src = src[1:]

	var schemaID int32
	if err := binary.Read(bytes.NewReader(src[:_SchemaIDSize]), binary.BigEndian, &schemaID); err != nil {
		return nil, nil, errors.Wrap(err, "failed to read avro schema id")
	}

	reqDecoder, err := deserializers.Get(ctx, schemaKind, int(schemaID))
	if err != nil {
		return nil, nil, err
	}

	return reqDecoder, src[_SchemaIDSize:], nil
}
