package database

import (
	"errors"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/Medzoner/gomedz/pkg/connector"
)

type DbMigration struct {
	RootPath     string
	DbInstance   connector.DbInstantiator
	MigrationDir string
}

// NewDbMigration is a function that returns a new DbMigration
func NewDbMigration(dbInstance connector.DbInstantiator, conf connector.Config) DbMigration {
	return DbMigration{
		DbInstance:   dbInstance,
		RootPath:     string(conf.RootPath),
		MigrationDir: string(conf.RootPath),
	}
}

const (
	Up   = "up"
	Down = "down"
)

// Migrate is a function that migrates down
func (d *DbMigration) Migrate(action string) error {
	db, err := d.getNewWithDatabaseInstance()
	if err != nil {
		return fmt.Errorf("database instantiate failed: %w", err)
	}

	switch action {
	case Up:
		if err = db.Up(); err != nil {
			return fmt.Errorf("database migration %s failed: %w", action, err)
		}
	case Down:
		if err = db.Down(); err != nil {
			return fmt.Errorf("database migration %s failed: %w", action, err)
		}
	default:
		return fmt.Errorf("database migration action failed: %w", err)
	}

	if err = d.checkMigrateErrors(err); err != nil {
		return fmt.Errorf("database checkMigrateErrors failed: %w", err)
	}

	log.Println("Database migrated ok: ", action)

	return nil
}

func (d *DbMigration) checkMigrateErrors(err error) error {
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("an error occurred while syncing the database: %s. %w", d.MigrationDir, err)
	}
	return nil
}

func (d *DbMigration) getNewWithDatabaseInstance() (*migrate.Migrate, error) {
	driver, err := d.DbInstance.GetDatabaseDriver()
	if err != nil {
		return nil, fmt.Errorf("database driver failed... %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", d.MigrationDir), d.DbInstance.GetDatabaseName(), driver)
	if err != nil {
		return nil, fmt.Errorf("database new instance failed... %w", err)
	}
	return m, nil
}
