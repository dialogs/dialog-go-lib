package avro

import (
	"context"
	"net"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/actgardner/gogen-avro/vm"
	"github.com/dialogs/dialog-go-lib/kafka/schemaregistry"
	"github.com/stretchr/testify/require"
)

func TestCacheGet(t *testing.T) {

	srClient := getSchemaRegistryClient(t)

	subject1 := "sbj" + strconv.FormatInt(time.Now().UnixNano(), 10)
	subject2 := "sbj" + strconv.FormatInt(time.Now().UnixNano(), 10)

	defer func() {
		_, err := srClient.DeleteSubject(context.Background(), subject1)
		require.NoError(t, err)

		_, err = srClient.DeleteSubject(context.Background(), subject2)
		require.NoError(t, err)
	}()

	newSchema1, err := srClient.RegisterNewSchema(
		context.Background(),
		subject1,
		`{"type": "record", "name": "TestObject1", "fields": [{"name": "Name", "type": "string"}]}`)
	require.NoError(t, err)

	newSchema2, err := srClient.RegisterNewSchema(
		context.Background(),
		subject2,
		`{"type": "record", "name": "TestObject2", "fields": [{"name": "Number", "type": "long"}]}`)
	require.NoError(t, err)

	cache := NewCache(srClient)

	{
		// test: not found
		d, err := cache.Get(context.Background(), KindUnknown, -1)
		require.EqualError(t, err, "failed to get schema for 'unknown': 404:40403 Schema not found")
		require.Nil(t, d)

		require.Equal(t,
			&Cache{
				schemaregistry: srClient,
				schemas: map[Kind]map[int]*Deserializer{
					KindUnknown: make(map[int]*Deserializer),
				},
			},
			cache)
	}

	for i := 0; i < 3; i++ {
		// any iterations for check cache

		// test: ok
		d1, err := cache.Get(context.Background(), KindUnknown, newSchema1.ID)
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
			d1.Instructions)

		d2, err := cache.Get(context.Background(), KindUnknown, newSchema2.ID)
		require.NoError(t, err)

		require.Equal(t,
			&Cache{
				schemaregistry: srClient,
				schemas: map[Kind]map[int]*Deserializer{
					KindUnknown: {
						newSchema1.ID: d1,
						newSchema2.ID: d2,
					},
				},
			},
			cache)
	}
}

func getSchemaRegistryClient(t *testing.T) *schemaregistry.Client {
	t.Helper()

	addr := (&url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort("localhost", "8081"),
	}).String()

	cfg := &schemaregistry.ConfigMock{
		Url:       addr,
		Timeout:   time.Second,
		Transport: nil,
	}
	client, err := schemaregistry.NewClient(cfg)
	require.NoError(t, err)

	return client
}
