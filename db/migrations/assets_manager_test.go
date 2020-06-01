package migrations

import (
	"io/ioutil"
	"testing"

	"github.com/dialogs/dialog-go-lib/db/migrations/test/esc"
	"github.com/stretchr/testify/require"
)

func TestManager(t *testing.T) {

	list, err := GetFilesList(esc.Assets, esc.DirName)
	require.NoError(t, err)
	require.Contains(t, list, "1_example.down.sql")
	require.Contains(t, list, "1_example.up.sql")

	fn := NewFileReaderHandler(esc.Assets, esc.DirName)
	data, err := fn("1_example.down.sql")
	require.NoError(t, err)

	require.Equal(t, "DROP TABLE example;", string(data))

}

func TestDriver(t *testing.T) {

	// test: fail
	driver, err := NewAssetsDriver(esc.Assets, esc.DirName+"1", GetFilesList)
	require.EqualError(t,
		err,
		"files for migrations assets driver: failed to open directory '/assets1': file does not exist")
	require.Nil(t, driver)

	// test: ok
	driver, err = NewAssetsDriver(esc.Assets, esc.DirName, GetFilesList)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, driver.Close())
	}()

	version, err := driver.First()
	require.NoError(t, err)

	{
		// test: up
		reader, identifier, err := driver.ReadUp(version)
		require.NoError(t, err)
		require.Equal(t, "example", identifier)

		b, err := ioutil.ReadAll(reader)
		require.NoError(t, err)
		require.Equal(t,
			"CREATE TABLE example\n(\n    id uuid NOT NULL PRIMARY KEY,\n    value varchar(1) NOT NULL\n);\n",
			string(b))
	}

	{
		// test: down
		reader, identifier, err := driver.ReadDown(version)
		require.NoError(t, err)
		require.Equal(t, "example", identifier)

		b, err := ioutil.ReadAll(reader)
		require.NoError(t, err)
		require.Equal(t,
			"DROP TABLE example;",
			string(b))
	}
}
