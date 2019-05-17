package migrations

import (
	"database/sql"
	"testing"

	"github.com/dialogs/dialog-go-lib/db/migrations/test"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestNewMigrate(t *testing.T) {

	db := getDb(t)

	const DbURL = "postgres://postgres@localhost:5432/example?sslmode=disable"

	m, err := NewMigrate(test.Assets, test.DirName, DbURL, GetFilesList)
	require.NoError(t, err)

	tableIsNotExist(t, db)

	defer func() {
		require.NoError(t, m.Down())
		tableIsNotExist(t, db)
	}()

	require.NoError(t, m.Up())
	tableIsExist(t, db)
}

func TestNewMigrateWithConn(t *testing.T) {

	db := getDb(t)

	m, err := NewMigrateWithConn(test.Assets, test.DirName, db, GetFilesList)
	require.NoError(t, err)

	tableIsNotExist(t, db)

	defer func() {
		require.NoError(t, m.Down())
		tableIsNotExist(t, db)
	}()

	require.NoError(t, m.Up())
	tableIsExist(t, db)
}

func tableIsExist(t *testing.T, db *sql.DB) {
	t.Helper()

	rows, err := db.Query("select * from example")
	require.NoError(t, err)

	cols, err := rows.Columns()
	require.NoError(t, err)
	require.Equal(t, []string{
		"id",
		"value",
	}, cols)
}

func tableIsNotExist(t *testing.T, db *sql.DB) {
	t.Helper()

	rows, err := db.Query("select * from example")
	e, ok := err.(*pq.Error)
	require.True(t, ok)
	require.Equal(t, "undefined_table", e.Code.Name())
	require.Equal(t, `relation "example" does not exist`, e.Message)

	require.Nil(t, rows)
}

func getDb(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("postgres", "host=localhost dbname=example user=postgres sslmode=disable")
	require.NoError(t, err)

	return db
}
