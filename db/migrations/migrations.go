package migrations

import (
	"database/sql"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/pkg/errors"
	pkgerr "github.com/pkg/errors"
)

func RegisterPostgresDriver() {
	// Stub for call postgres.init()
	// Fix error: 'database driver: unknown driver postgres (forgotten import?)'
	if postgres.DefaultMigrationsTable == "" {
		panic("fatal")
	}
}

// A Migrate is a wrapper
type Migrate migrate.Migrate

// NewMigrate create a new migration driver
func NewMigrate(dbURL string, fs http.FileSystem, dirName string, getAssets FilesList) (*Migrate, error) {

	assetsDriver, err := NewAssetsDriver(fs, dirName, getAssets)
	if err != nil {
		return nil, pkgerr.Wrap(err, "failed to create migrations data driver")
	}

	m, err := migrate.NewWithSourceInstance("go-bindata", assetsDriver, dbURL)
	if err != nil {
		return nil, pkgerr.Wrap(err, "failed to create migrations instance")
	}

	return (*Migrate)(m), nil
}

// NewMigrateWithConn create a new migration driver by existing database connection
func NewMigrateWithConn(db *sql.DB, fs http.FileSystem, dirName string, getAssets FilesList) (*Migrate, error) {

	assetsDriver, err := NewAssetsDriver(fs, dirName, getAssets)
	if err != nil {
		return nil, pkgerr.Wrap(err, "new assets driver")
	}

	dbDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, pkgerr.Wrap(err, "failed to create migrations data driver")
	}

	m, err := migrate.NewWithInstance("go-bindata", assetsDriver, "postgres", dbDriver)
	if err != nil {
		return nil, pkgerr.Wrap(err, "failed to create migrations instance")
	}

	return (*Migrate)(m), nil
}

func NewMigrateFromCustomSource(dbURL string, assetsNames func() []string, getAsset func(name string) ([]byte, error)) (*Migrate, error) {

	s := bindata.Resource(assetsNames(),
		func(name string) (data []byte, err error) {
			data, err = getAsset(name)
			err = errors.Wrap(err, "failed to get migration data: "+name)
			return
		})

	d, err := bindata.WithInstance(s)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create migrations data driver")
	}

	m, err := migrate.NewWithSourceInstance("migrations", d, dbURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create migrations instance")
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

// Steps looks at the currently active migration version.
// It will migrate up if n > 0, and down if n < 0.
func (m *Migrate) Steps(n int) error {

	native := (*migrate.Migrate)(m)
	err := native.Steps(n)

	if err != nil && err != migrate.ErrNoChange {
		return pkgerr.Wrap(err, "failed to do steps")
	}

	return nil
}
