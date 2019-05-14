package migrations

import (
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	pkgerr "github.com/pkg/errors"
)

func init() {
	// Stub for call postgres.init()
	// Fix error: 'database driver: unknown driver postgres (forgotten import?)'
	if (&postgres.Postgres{}) != nil {
	}
}

// A Migrate is a wrapper
type Migrate migrate.Migrate

// NewMigrate create a new migration driver
func NewMigrate(fs http.FileSystem, dirName, dbURL string, getAssets FilesList) (*Migrate, error) {

	d, err := NewAssetsDriver(fs, dirName, getAssets)
	if err != nil {
		return nil, pkgerr.Wrap(err, "new assets driver")
	}

	m, err := migrate.NewWithSourceInstance("go-bindata", d, dbURL)
	if err != nil {
		return nil, pkgerr.Wrap(err, "new migrator")
	}

	return (*Migrate)(m), nil
}

// Up applies new migrations
func (m *Migrate) Up() error {

	native := (*migrate.Migrate)(m)
	err := native.Up()

	if err != nil && err != migrate.ErrNoChange {
		return pkgerr.Wrap(err, "migrations: up")
	}

	return nil
}

// Down applies all down migrations
func (m *Migrate) Down() error {

	native := (*migrate.Migrate)(m)
	err := native.Down()

	if err != nil && err != migrate.ErrNoChange {
		return pkgerr.Wrap(err, "migrations: down")
	}

	return nil
}
