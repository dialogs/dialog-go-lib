package schemaregistry

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/dialogs/dialog-go-lib/serde/avro"

	"github.com/actgardner/gogen-avro/vm"
	"github.com/bluele/gcache"
	"github.com/stretchr/testify/require"
)

func TestGCacheLoaderFunc(t *testing.T) {

	subject := "sbj" + strconv.FormatInt(time.Now().UnixNano(), 10)

	ctx := context.Background()
	cfg := NewConfigMock(BaseUrl, time.Second*5, nil)

	client, err := NewClient(cfg)
	require.NoError(t, err)

	defer func() {
		// test: remove subject
		_, err := client.DeleteSubject(ctx, subject)
		require.NoError(t, err)

		{
			// check after remove
			res, err := client.GetSubjectList(ctx)
			require.NoError(t, err)
			require.Equal(t, ResGetSubjectList{}, res)
		}
	}()

	schema := `{"type": "record", "name": "TestObject", "fields": [{"name": "Name", "type": "string"}]}`

	// test: add new schema
	setRes, err := client.RegisterNewSchema(ctx, subject, schema)
	require.NoError(t, err)

	c := gcache.New(10).LoaderFunc(gcache.LoaderFunc(NewGCacheDeserializerLoaderFunc(client))).LRU().Build()

	{
		// test: ok
		val, err := c.Get(setRes.ID)
		require.NoError(t, err)
		require.Equal(t,
			[]vm.Instruction{
				{Op: 7, Operand: 2},
				{Op: 9, Operand: 0},
				{Op: 2, Operand: 0},
				{Op: 0, Operand: 8},
				{Op: 1, Operand: 8},
				{Op: 3, Operand: 65535},
				{Op: 8, Operand: 65535},
			},
			val.(*avro.Deserializer).Program.Instructions)
	}

	{
		// test: invalid key
		val, err := c.Get("a")
		require.EqualError(t, err, "invalid key type: string")
		require.Nil(t, val)
	}

	{
		// test: schema not found
		val, err := c.Get(9999)
		require.EqualError(t, err, "failed to get schema: 404:40403 Schema not found")
		require.Nil(t, val)
	}
}
