package database_test

import (
	"errors"
	"testing"

	"github.com/Medzoner/gomedz/pkg/connector"
	"github.com/Medzoner/medzoner-go/internal/database"
	migratedb "github.com/golang-migrate/migrate/v4/database"
	"github.com/jmoiron/sqlx"
	"gotest.tools/assert"
)

// failingDbInstantiator returns errors on GetDatabaseDriver
type failingDbInstantiator struct {
	driverErr error
}

func (f *failingDbInstantiator) GetConnection() *sqlx.DB { return nil }
func (f *failingDbInstantiator) CreateDatabase(_ string) {}
func (f *failingDbInstantiator) DropDatabase(_ string)   {}
func (f *failingDbInstantiator) GetDatabaseName() string { return "test" }
func (f *failingDbInstantiator) GetDatabaseDriver() (migratedb.Driver, error) {
	return nil, f.driverErr
}
func (f *failingDbInstantiator) Connect() *sqlx.DB { return nil }

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

	t.Run("Unit: test NewDbMigration sets fields correctly", func(t *testing.T) {
		conf := connector.Config{RootPath: "/app/migrations/"}
		dbMigration := database.NewDbMigration(
			&failingDbInstantiator{},
			conf,
		)

		assert.Equal(t, dbMigration.RootPath, "/app/migrations/")
		assert.Equal(t, dbMigration.MigrationDir, "/app/migrations/")
	})
}

func TestDbMigration_CheckMigrateErrors(t *testing.T) {
	t.Run("Unit: test checkMigrateErrors is called via Migrate with nil driver", func(t *testing.T) {
		// When GetDatabaseDriver returns nil driver (no error), getNewWithDatabaseInstance
		// will fail at migrate.NewWithDatabaseInstance with a nil driver
		dbMigration := database.NewDbMigration(
			&failingDbInstantiator{driverErr: nil},
			connector.Config{RootPath: "./nonexistent/"},
		)

		err := dbMigration.Migrate(database.Up)

		assert.ErrorContains(t, err, "database instantiate failed")
	})
}

