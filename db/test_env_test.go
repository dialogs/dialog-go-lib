package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHelper(t *testing.T) {

	conf := Config{
		Host:                   "localhost",
		Port:                   "5432",
		Name:                   "postgres",
		User:                   "postgres",
		Password:               "123",
		SslMode:                "disable",
		HealthCheckPeriod:      time.Second,
		MaxConnections:         10,
		StatementCacheCapacity: 10,
	}

	env, err := NewTestEnv(conf.ConnURLWithoutSchema())
	require.NoError(t, err)

	for _, testInfo := range []struct {
		Name string
		Func func(*testing.T, *TestEnv)
	}{
		{Name: "create/drop/exist", Func: testTestEnvSchemeCreateDropExist},
		{Name: "clear", Func: testTestEnvClearDB},
		{Name: "columns", Func: testTestEnvGetColumns},
	} {

		fn := func(*testing.T) { testInfo.Func(t, env) }
		if !t.Run(testInfo.Name, fn) {
			return
		}
	}
}

func testTestEnvSchemeCreateDropExist(t *testing.T, env *TestEnv) {

	Scheme := "test" + strconv.FormatInt(time.Now().UnixNano(), 10)

	require.Equal(t, sql.ErrNoRows, env.SchemeExists(Scheme))

	require.NoError(t, env.CreateScheme(Scheme))
	defer func() { require.NoError(t, env.DropScheme(Scheme)) }()

	require.NoError(t, env.SchemeExists(Scheme))
}

func testTestEnvClearDB(t *testing.T, env *TestEnv) {

	Scheme := "test" + strconv.FormatInt(time.Now().UnixNano(), 10)
	var TablesNames = []string{Scheme + ".temp1", Scheme + ".temp2"}

	fnCountRows := func(tablesNames ...string) (total int) {
		var count int32
		for _, table := range tablesNames {
			err := env.Conn().QueryRow(fmt.Sprintf("select count(*) from %s", table)).Scan(&count)
			require.NoError(t, err)
			total += int(count)
		}
		return
	}

	{
		// create environment
		require.NoError(t, env.CreateScheme(Scheme))
		defer func() { require.NoError(t, env.DropScheme(Scheme)) }()

		for _, table := range TablesNames {
			_, err := env.Conn().Exec(fmt.Sprintf("create table %s (id integer)", table))
			require.NoError(t, err)

			_, err = env.Conn().Exec(fmt.Sprintf("insert into %s (id) values(1)", table))
			require.NoError(t, err)
		}
	}

	// before
	require.Equal(t, len(TablesNames), fnCountRows(TablesNames...))

	query := make([]string, 0)
	for _, n := range TablesNames {
		query = append(query, "delete from "+n)
	}
	require.NoError(t, env.ClearDB(query...))

	// after
	require.Equal(t, 0, fnCountRows(TablesNames...))
}

func testTestEnvGetColumns(t *testing.T, env *TestEnv) {

	const Table = "temp1"
	Scheme := "test" + strconv.FormatInt(time.Now().UnixNano(), 10)

	{
		// create environment
		require.NoError(t, env.CreateScheme(Scheme))
		defer func() { require.NoError(t, env.DropScheme(Scheme)) }()

		_, err := env.Conn().Exec(fmt.Sprintf("create table %s (id integer, val text)", Scheme+"."+Table))
		require.NoError(t, err)
	}

	columns, err := env.GetTableColumns(Scheme, Table)
	require.NoError(t, err)
	require.Equal(t,
		map[string]string{
			"id":  "integer",
			"val": "text",
		},
		columns)
}
