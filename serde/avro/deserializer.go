package avro

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/actgardner/gogen-avro/compiler"
	"github.com/actgardner/gogen-avro/vm"
	"github.com/actgardner/gogen-avro/vm/types"
	"github.com/pkg/errors"
)

type Deserializer struct {
	*vm.Program
}

func NewDeserializer(fromSchema, toSchema string) (*Deserializer, error) {

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

func ParseForDeserializer(src []byte) (schemaID int32, payload []byte, _ error) {

	const SchemaIDSize = 4

	if len(src) < SchemaIDSize+1 { // +1 magic byte
		return -1, nil, fmt.Errorf("invalid incoming avro request: %#v", src)
	}

	if src[0] != 0x00 {
		return -1, nil, fmt.Errorf("invalid magic byte in incoming avro request: %#v", src)
	}

	src = src[1:] // exclude magic byte

	if err := binary.Read(bytes.NewReader(src[:SchemaIDSize]), binary.BigEndian, &schemaID); err != nil {
		return -1, nil, errors.Wrap(err, "failed to read avro schema id")
	}

	payload = src[SchemaIDSize:]

	return
}
