package registry

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const BaseUrl = "http://localhost:8081"

func TestConfig(t *testing.T) {

	ctx := context.Background()

	r, err := New(BaseUrl, time.Second, nil)
	require.NoError(t, err)

	defer func() {
		// set default value
		require.NoError(t,
			r.SetConfig(ctx, "", &ReqConfig{Compatibility: "BACKWARD"}))
	}()

	require.EqualError(t,
		r.SetConfig(ctx, "", &ReqConfig{Compatibility: "unknown"}),
		"422:42203 Invalid compatibility level. Valid values are none, backward, forward and full")

	{
		// test: default settings
		cfg, err := r.GetConfig(ctx, "")
		require.NoError(t, err)
		require.Equal(t, &ResConfig{Compatibility: "BACKWARD"}, cfg)
	}

	// test: set new value
	require.NoError(t,
		r.SetConfig(ctx, "", &ReqConfig{Compatibility: "full"}))

	{
		// test: get new value
		cfg, err := r.GetConfig(ctx, "")
		require.NoError(t, err)
		require.Equal(t, &ResConfig{Compatibility: "FULL"}, cfg)
	}
}

func TestSubject(t *testing.T) {

	subject := "sbj" + strconv.FormatInt(time.Now().UnixNano(), 10)

	ctx := context.Background()

	r, err := New(BaseUrl, time.Second, nil)
	require.NoError(t, err)

	defer func() {
		// test: remove subject
		_, err := r.DeleteSubject(ctx, subject)
		require.NoError(t, err)

		{
			// check after remove
			res, err := r.GetSubjectList(ctx)
			require.NoError(t, err)
			require.Equal(t, ResGetSubjectList{}, res)
		}
	}()

	{
		// test: check subject before
		res, err := r.GetSubjectList(ctx)
		require.NoError(t, err)
		require.Equal(t, ResGetSubjectList{}, res)
	}

	schema := `{"type": "record", "name": "TestObject", "fields": [{"name": "Name", "type": "string"}]}`
	schemaShort := strings.ReplaceAll(schema, " ", "")

	{
		// test: check schema before
		setRes, err := r.CheckSubject(ctx, subject, schema)
		require.EqualError(t, err, "404:40401 Subject not found.")
		require.Nil(t, setRes)
	}

	// test: add new schema
	setRes, err := r.RegisterNewSchema(ctx, subject, schema)
	require.NoError(t, err)
	require.True(t, setRes.ID >= 0)
	require.Equal(t, &ResRegisterNewSchema{ID: setRes.ID}, setRes)

	{
		// test: get subjects list
		res, err := r.GetSubjectList(ctx)
		require.NoError(t, err)
		require.Equal(t, ResGetSubjectList{subject}, res)
	}

	{
		// test: get latest subject version
		res, err := r.GetSubjectVersion(ctx, subject, -1)
		require.NoError(t, err)
		require.Equal(t,
			&ResGetSubjectVersion{
				Name:    "",
				Version: 1,
				Schema:  schemaShort,
			},
			res)
	}

	{
		// test: check subject
		res, err := r.CheckSubject(ctx, subject, schema)
		require.NoError(t, err)
		require.Equal(t,
			&ResCheckSubject{
				Subject: subject,
				ID:      setRes.ID,
				Version: 1,
				Schema:  schemaShort,
			},
			res)
	}

	{
		// test: get schema
		res, err := r.GetSchema(ctx, setRes.ID)
		require.NoError(t, err)
		require.Equal(t,
			&ResSchema{
				Schema: schemaShort,
			},
			res)
	}
}

func TestSubjectVersionsList(t *testing.T) {

	ctx := context.Background()
	subject := "sbj" + strconv.FormatInt(time.Now().UnixNano(), 10)

	r, err := New(BaseUrl, time.Second, nil)
	require.NoError(t, err)

	defer func() {
		// test: remove subject
		_, err := r.DeleteSubject(ctx, subject)
		require.NoError(t, err)
	}()

	schema1 := `{"type": "record", "name": "TestObject", "fields": [{"name": "Name", "type": "string"}]}`
	schema2 := `{"type": "record", "name": "TestObject", "fields": [{"name": "FistName", "type": "string"}]}`
	schema2Short := strings.ReplaceAll(schema2, " ", "")

	setRes1, err := r.RegisterNewSchema(ctx, subject, schema1)
	require.NoError(t, err)
	require.True(t, setRes1.ID >= 0)

	// test: set config for subject
	require.NoError(t,
		r.SetConfig(ctx, subject, &ReqConfig{Compatibility: "NONE"}))

	{
		// test: check config for subject
		cfg, err := r.GetConfig(ctx, subject)
		require.NoError(t, err)
		require.Equal(t, &ResConfig{Compatibility: "NONE"}, cfg)
	}

	{
		// test: get subject versions list
		res, err := r.GetSubjectVersionsList(ctx, subject)
		require.NoError(t, err)
		require.Equal(t, ResGetSubjectVersionsList{1}, res)
	}

	// test: add new schema version
	setRes2, err := r.RegisterNewSchema(ctx, subject, schema2)
	require.NoError(t, err)
	require.True(t, setRes2.ID >= 0)

	{
		// test: get latest schema version
		res, err := r.GetSubjectVersion(ctx, subject, -1)
		require.NoError(t, err)
		require.Equal(t,
			&ResGetSubjectVersion{
				Name:    "",
				Version: 2,
				Schema:  schema2Short,
			},
			res)
	}

	{
		// test: get subject versions list
		res, err := r.GetSubjectVersionsList(ctx, subject)
		require.NoError(t, err)
		require.Equal(t, ResGetSubjectVersionsList{1, 2}, res)
	}

	{
		// test: remove subject version
		res, err := r.DeleteSubjectVersion(ctx, subject, 1)
		require.NoError(t, err)
		require.Equal(t, ResDeleteSubjectVersion(1), res)
	}
}
