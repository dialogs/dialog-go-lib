package avro

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

type ISerializer interface {
	Serialize(io.Writer) error
}

func Encode(schemaID int32, obj ISerializer) ([]byte, error) {

	buf := bytes.NewBuffer(nil)
	if err := buf.WriteByte(0x00); err != nil {
		return nil, errors.Wrap(err, "failed to write magic byte")
	}

	if err := binary.Write(buf, binary.BigEndian, schemaID); err != nil {
		return nil, errors.Wrap(err, "failed to encode outgoing schema number")
	}

	if err := obj.Serialize(buf); err != nil {
		return nil, errors.Wrap(err, "failed to encode outgoing object")
	}

	return buf.Bytes(), nil
}
