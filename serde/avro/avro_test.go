package avro

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

type Encoder struct {
	payload string
}

func NewEncoder(payload string) *Encoder {
	return &Encoder{
		payload: payload,
	}
}

func (e *Encoder) Serialize(w io.Writer) (err error) {
	_, err = w.Write([]byte(e.payload))
	return
}

func TestEncodeDecode(t *testing.T) {

	const (
		SchemaID int32  = 123
		Payload  string = "abc"
	)

	enc, err := Encode(SchemaID, NewEncoder(Payload))
	require.NoError(t, err)
	require.NotNil(t, enc)

	{
		// test: ok
		schema, payload, err := ParseForDeserializer(enc)
		require.NoError(t, err)
		require.Equal(t, SchemaID, schema)
		require.Equal(t, []byte(Payload), payload)
	}

	{
		// test: invalid request
		schema, payload, err := ParseForDeserializer(make([]byte, 4))
		require.EqualError(t, err, "invalid incoming avro request: []byte{0x0, 0x0, 0x0, 0x0}")
		require.Equal(t, int32(-1), schema)
		require.Nil(t, payload)
	}

	{
		// test: invalid magic byte
		schema, payload, err := ParseForDeserializer(append([]byte{0x01}, enc...))
		require.EqualError(t, err, "invalid magic byte in incoming avro request: []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x7b, 0x61, 0x62, 0x63}")
		require.Equal(t, int32(-1), schema)
		require.Nil(t, payload)
	}
}
