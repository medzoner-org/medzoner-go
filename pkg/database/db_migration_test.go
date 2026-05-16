package database_test

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Medzoner/gomedz/pkg/connector"
	migratedb "github.com/golang-migrate/migrate/v4/database"
	"github.com/jmoiron/sqlx"
	"gotest.tools/assert"
	"github.com/Medzoner/medzoner-go/pkg/database"
)

// failingDbInstantiator returns errors on GetDatabaseDriver
type failingDbInstantiator struct {
	driverErr error
	driver    migratedb.Driver
}

func (f *failingDbInstantiator) GetConnection() *sqlx.DB { return nil }
func (f *failingDbInstantiator) CreateDatabase(_ string) {}
func (f *failingDbInstantiator) DropDatabase(_ string)   {}
func (f *failingDbInstantiator) GetDatabaseName() string { return "test" }
func (f *failingDbInstantiator) GetDatabaseDriver() (migratedb.Driver, error) {
	if f.driverErr != nil {
		return nil, f.driverErr
	}
	return f.driver, nil
}
func (f *failingDbInstantiator) Connect() *sqlx.DB { return nil }

// fakeDriver implements migrate/database.Driver
type fakeDriver struct {
	version int
	dirty   bool
}

func (f *fakeDriver) Open(_ string) (migratedb.Driver, error) { return f, nil }
func (f *fakeDriver) Close() error                            { return nil }
func (f *fakeDriver) Lock() error                             { return nil }
func (f *fakeDriver) Unlock() error                           { return nil }
func (f *fakeDriver) Run(_ io.Reader) error                   { return nil }
func (f *fakeDriver) SetVersion(version int, dirty bool) error {
	f.version = version
	f.dirty = dirty
	return nil
}
func (f *fakeDriver) Version() (int, bool, error) { return f.version, f.dirty, nil }
func (f *fakeDriver) Drop() error                 { return nil }

func createMigrationFiles(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	migDir := filepath.Join(tmpDir, "migrations")
	err := os.MkdirAll(migDir, 0o755)
	assert.NilError(t, err)

	err = os.WriteFile(filepath.Join(migDir, "000001_init.up.sql"), []byte("CREATE TABLE test (id INT);"), 0o644)
	assert.NilError(t, err)
	err = os.WriteFile(filepath.Join(migDir, "000001_init.down.sql"), []byte("DROP TABLE test;"), 0o644)
	assert.NilError(t, err)

	return migDir
}

func TestDbMigration_Migrate(t *testing.T) {
	t.Run("Unit: test Migrate up with driver error", func(t *testing.T) {
		dbMigration := database.NewDbMigration(
			&failingDbInstantiator{driverErr: errors.New("driver failure")},
			connector.Config{RootPath: "./"},
		)

		err := dbMigration.Migrate(database.Up)

		assert.ErrorContains(t, err, "database instantiate failed")
		assert.ErrorContains(t, err, "driver failure")
	})

	t.Run("Unit: test Migrate down with driver error", func(t *testing.T) {
		dbMigration := database.NewDbMigration(
			&failingDbInstantiator{driverErr: errors.New("driver failure")},
			connector.Config{RootPath: "./"},
		)

		err := dbMigration.Migrate(database.Down)

		assert.ErrorContains(t, err, "database instantiate failed")
	})

	t.Run("Unit: test Migrate with invalid action and driver error", func(t *testing.T) {
		dbMigration := database.NewDbMigration(
			&failingDbInstantiator{driverErr: errors.New("driver failure")},
			connector.Config{RootPath: "./"},
		)

		err := dbMigration.Migrate("invalid")

		assert.ErrorContains(t, err, "database instantiate failed")
	})

	t.Run("Unit: test NewDbMigration sets MigrationDir correctly", func(t *testing.T) {
		conf := connector.Config{RootPath: "/app/"}
		dbMigration := database.NewDbMigration(
			&failingDbInstantiator{},
			conf,
		)

		assert.Equal(t, dbMigration.MigrationDir, "/app/migrations/")
	})

	t.Run("Unit: test Migrate up success with fake driver", func(t *testing.T) {
		migDir := createMigrationFiles(t)

		dbMigration := database.DbMigration{
			DbInstance:   &failingDbInstantiator{driver: &fakeDriver{version: -1}},
			MigrationDir: migDir,
		}

		err := dbMigration.Migrate(database.Up)

		assert.NilError(t, err)
	})

	t.Run("Unit: test Migrate down success with fake driver", func(t *testing.T) {
		migDir := createMigrationFiles(t)

		dbMigration := database.DbMigration{
			DbInstance:   &failingDbInstantiator{driver: &fakeDriver{version: 1}},
			MigrationDir: migDir,
		}

		err := dbMigration.Migrate(database.Down)

		assert.NilError(t, err)
	})

	t.Run("Unit: test Migrate with invalid action and valid driver", func(t *testing.T) {
		migDir := createMigrationFiles(t)

		dbMigration := database.DbMigration{
			DbInstance:   &failingDbInstantiator{driver: &fakeDriver{version: -1}},
			MigrationDir: migDir,
		}

		err := dbMigration.Migrate("invalid")

		assert.ErrorContains(t, err, "database migration action failed")
		assert.ErrorContains(t, err, "unsupported action")
	})

	t.Run("Unit: test Migrate up no change is not an error", func(t *testing.T) {
		migDir := createMigrationFiles(t)

		dbMigration := database.DbMigration{
			DbInstance:   &failingDbInstantiator{driver: &fakeDriver{version: 1}},
			MigrationDir: migDir,
		}

		err := dbMigration.Migrate(database.Up)

		assert.NilError(t, err)
	})
}

func TestDbMigration_CheckMigrateErrors(t *testing.T) {
	t.Run("Unit: test instantiate with nil driver", func(t *testing.T) {
		dbMigration := database.NewDbMigration(
			&failingDbInstantiator{driverErr: nil},
			connector.Config{RootPath: "./nonexistent/"},
		)

		err := dbMigration.Migrate(database.Up)

		assert.ErrorContains(t, err, "database instantiate failed")
	})
}
