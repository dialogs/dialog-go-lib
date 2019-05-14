package migrations

import (
	"testing"

	"github.com/dialogs/dialog-go-lib/db/migrations/test"
	"github.com/stretchr/testify/require"
)

func TestMigrations(t *testing.T) {

	const DbURL = "postgres://postgres@localhost:5432/example?sslmode=disable"

	m, err := NewMigrate(test.Assets, test.DirName, DbURL, GetFilesList)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, m.Down())
	}()

	require.NoError(t, m.Up())

}
