package migrations

import (
	"database/sql"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	pkgerr "github.com/pkg/errors"
)

func init() {
	// Stub for call postgres.init()
	// Fix error: 'database driver: unknown driver postgres (forgotten import?)'
	if (&postgres.Postgres{}) == nil {
		panic("fatal")
	}
}

// A Migrate is a wrapper
type Migrate migrate.Migrate

// NewMigrate create a new migration driver
func NewMigrate(fs http.FileSystem, dirName, dbURL string, getAssets FilesList) (*Migrate, error) {

	assetsDriver, err := NewAssetsDriver(fs, dirName, getAssets)
	if err != nil {
		return nil, pkgerr.Wrap(err, "new assets driver")
	}

	m, err := migrate.NewWithSourceInstance("go-bindata", assetsDriver, dbURL)
	if err != nil {
		return nil, pkgerr.Wrap(err, "new migrator")
	}

	return (*Migrate)(m), nil
}

// NewMigrateWithConn create a new migration driver by existing database connection
func NewMigrateWithConn(fs http.FileSystem, dirName string, db *sql.DB, getAssets FilesList) (*Migrate, error) {

	assetsDriver, err := NewAssetsDriver(fs, dirName, getAssets)
	if err != nil {
		return nil, pkgerr.Wrap(err, "new assets driver")
	}

	dbDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, pkgerr.Wrap(err, "new database driver")
	}

	m, err := migrate.NewWithInstance("go-bindata", assetsDriver, "postgres", dbDriver)
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
