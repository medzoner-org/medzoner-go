package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/Medzoner/gomedz/pkg/connector"
	"github.com/Medzoner/gomedz/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// DbMigration handles database schema migrations.
type DbMigration struct {
	DbInstance   connector.DbInstantiator
	MigrationDir string
}

// NewDbMigration creates a new DbMigration with the migration directory derived from config RootPath.
func NewDbMigration(dbInstance connector.DbInstantiator, conf connector.Config) DbMigration {
	return DbMigration{
		DbInstance:   dbInstance,
		MigrationDir: string(conf.RootPath) + "migrations/",
	}
}

const (
	// Up migrates the database to the latest version.
	Up = "up"
	// Down rolls back the last database migration.
	Down = "down"
)

// Migrate runs a migration action (Up or Down).
func (d *DbMigration) Migrate(action string) error {
	db, err := d.newMigrateInstance()
	if err != nil {
		return fmt.Errorf("database instantiate failed: %w", err)
	}
	if err = d.runAction(db, action); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info(context.TODO(), fmt.Sprintf("database migration %s: no change", action))
			return nil
		}
		return err
	}
	logger.Info(context.TODO(), fmt.Sprintf("database migrated ok: %s", action))
	return nil
}
func (d *DbMigration) runAction(db *migrate.Migrate, action string) error {
	switch action {
	case Up:
		if err := db.Up(); err != nil {
			return fmt.Errorf("database migration %s failed: %w", action, err)
		}
	case Down:
		if err := db.Down(); err != nil {
			return fmt.Errorf("database migration %s failed: %w", action, err)
		}
	default:
		return fmt.Errorf("database migration action failed: unsupported action %q", action)
	}
	return nil
}
func (d *DbMigration) newMigrateInstance() (*migrate.Migrate, error) {
	driver, err := d.DbInstance.GetDatabaseDriver()
	if err != nil {
		return nil, fmt.Errorf("database driver failed: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", d.MigrationDir), d.DbInstance.GetDatabaseName(), driver)
	if err != nil {
		return nil, fmt.Errorf("database new instance failed: %w", err)
	}
	return m, nil
}
