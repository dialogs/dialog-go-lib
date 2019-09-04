package avro

import (
	"strconv"
	"testing"

	"github.com/dialogs/dialog-go-lib/serde"
	"github.com/stretchr/testify/require"
)

func TestAvro(t *testing.T) {

	newVal := func(i int) map[string]interface{} {
		strI := strconv.Itoa(i)

		return map[string]interface{}{
			"f_string":     strI,
			"f_int64":      int64(i),
			"f_string_map": map[string]interface{}{strI: strI},
		}
	}

	schema := `{
		"type": "record",
		"name": "Object",
		"fields": [
			{"name": "f_string", "type": "string"},
			{"name": "f_int64", "type": "long"},
			{"name": "f_string_map", "type": "map", "values": "string"}
		]}`

	var (
		s   serde.Serde
		err error
	)
	s, err = New(schema)
	require.NoError(t, err)

	encVal, err := s.Encode(newVal(2))
	require.NoError(t, err)

	decVal, err := s.Decode(encVal)
	require.NoError(t, err)

	require.Equal(t, newVal(2), decVal)
}

func TestInvalidSchema(t *testing.T) {

	schema := `{}`

	s, err := New(schema)
	require.EqualError(t, err, "failed to read avro schema: missing type: map[]")
	require.Nil(t, s)
}

func TestFailedEncode(t *testing.T) {

	schema := `{
		"type": "record",
		"name": "Object",
		"fields": [
			{"name": "f_string", "type": "string"},
			{"name": "f_int64", "type": "long"},
			{"name": "f_string_map", "type": "map", "values": "string"}
		]}`

	s, err := New(schema)
	require.NoError(t, err)

	res, err := s.Encode([]int{1, 2, 3})
	require.Nil(t, res)
	require.EqualError(t,
		err,
		"cannot encode binary record \"Object\": expected map[string]interface{}; received: []int")
}

func TestFailedDecode(t *testing.T) {

	schema := `{
		"type": "record",
		"name": "Object",
		"fields": [
			{"name": "f_string", "type": "string"},
			{"name": "f_int64", "type": "long"},
			{"name": "f_string_map", "type": "map", "values": "string"}
		]}`

	s, err := New(schema)
	require.NoError(t, err)

	res, err := s.Decode([]byte{1, 2, 3})
	require.Nil(t, res)
	require.EqualError(t,
		err,
		"cannot decode binary record \"Object\" field \"f_string\": cannot decode binary string: cannot decode binary bytes: negative size: -1")
}
