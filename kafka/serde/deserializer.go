package serde

import (
	"io"

	"github.com/actgardner/gogen-avro/compiler"
	"github.com/actgardner/gogen-avro/vm"
	"github.com/actgardner/gogen-avro/vm/types"
)

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
