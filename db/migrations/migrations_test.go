package migrations

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/dialogs/dialog-go-lib/db"
	"github.com/dialogs/dialog-go-lib/db/migrations/test/esc"
	"github.com/dialogs/dialog-go-lib/db/migrations/test/gobindata"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/stretchr/testify/require"
)

func TestMigrate(t *testing.T) {

	conf := db.Config{
		Host:                   "localhost",
		Port:                   "5432",
		Name:                   "postgres",
		Scheme:                 "test" + strconv.FormatInt(time.Now().UnixNano(), 10),
		User:                   "postgres",
		Password:               "123",
		SslMode:                "disable",
		HealthCheckPeriod:      time.Second,
		MaxConnections:         10,
		StatementCacheCapacity: 10,
	}

	env, err := db.NewTestEnv(conf.ConnURLWithoutSchema())
	require.NoError(t, err)

	for i, testInfo := range []struct {
		Name string
		Func func(db.Config) (*Migrate, error)
	}{
		{Name: "new migrate", Func: newMigrate},
		{Name: "migrate with conn", Func: newMigrateWithConn},
		{Name: "migrate from custom source", Func: newMigrateFromCustomSource},
	} {

		testDesc := fmt.Sprintf("#%d", i)

		fn := func(*testing.T) {
			require.NoError(t, env.CreateScheme(conf.Scheme))
			defer func() { require.NoError(t, env.DropScheme(conf.Scheme), testDesc) }()

			m, err := testInfo.Func(conf)
			require.NoError(t, err, testDesc)

			applyMigrations(t, testDesc, env, conf.Scheme, m.Up, m.Down)
			applyMigrations(t, testDesc, env, conf.Scheme, func() error { return m.Steps(1) }, func() error { return m.Steps(-1) })
		}

		if !t.Run(testInfo.Name, fn) {
			return
		}
	}
}

func newMigrate(conf db.Config) (m *Migrate, err error) {
	m, err = NewMigrate(conf.ConnURL(), esc.Assets, esc.DirName, GetFilesList)
	return
}

func newMigrateWithConn(conf db.Config) (m *Migrate, err error) {

	connConf, err := pgx.ParseConfig(conf.ConnURL())
	if err != nil {
		return
	}

	db := stdlib.OpenDB(*connConf)

	m, err = NewMigrateWithConn(db, esc.Assets, esc.DirName, GetFilesList)
	return
}

func newMigrateFromCustomSource(conf db.Config) (m *Migrate, err error) {
	m, err = NewMigrateFromCustomSource(conf.ConnURL(), gobindata.AssetNames, gobindata.Asset)
	return
}

func applyMigrations(t *testing.T, testDesc string, env *db.TestEnv, scheme string, up, down func() error) {

	const Table = "example"
	var EmptyTable = map[string]string{}

	columns, err := env.GetTableColumns(scheme, Table)
	require.NoError(t, err, testDesc)
	require.Equal(t, EmptyTable, columns, testDesc)

	defer func() {
		require.NoError(t, down(), testDesc)
		columns, err := env.GetTableColumns(scheme, Table)
		require.NoError(t, err, testDesc)
		require.Equal(t, EmptyTable, columns, testDesc)
	}()

	require.NoError(t, up(), testDesc)
	columns, err = env.GetTableColumns(scheme, Table)
	require.NoError(t, err, testDesc)
	require.Equal(t, map[string]string{"id": "uuid", "value": "character varying"}, columns, testDesc)
}
