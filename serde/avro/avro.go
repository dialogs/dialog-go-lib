package avro

import (
	goavro "github.com/linkedin/goavro/v2"
	"github.com/pkg/errors"
)

type Avro struct {
	codec *goavro.Codec
}

func New(schema string) (*Avro, error) {

	codec, err := goavro.NewCodec(schema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read avro schema")
	}

	return &Avro{
		codec: codec,
	}, nil
}

func (a *Avro) Encode(val interface{}) ([]byte, error) {

	encData, err := a.codec.BinaryFromNative(nil, val)
	if err != nil {
		return nil, err
	}

	return encData, nil
}

func (a *Avro) Decode(data []byte) (interface{}, error) {

	val, _, err := a.codec.NativeFromBinary(data)
	if err != nil {
		return nil, err
	}

	return val, nil
}
