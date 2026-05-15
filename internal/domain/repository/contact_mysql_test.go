package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Medzoner/gomedz/pkg/logger"
	"github.com/Medzoner/gomedz/pkg/observability"
	"github.com/Medzoner/medzoner-go/internal/domain/repository"
	"github.com/Medzoner/medzoner-go/internal/entity"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v1"
	"gotest.tools/assert"
)

func init() {
	l, err := logger.NewLogger(logger.Config{Level: "debug"})
	if err != nil {
		panic(err)
	}
	_, _ = observability.NewTelemetry(context.Background(), observability.Config{}, l)
}

// fakeDbInstantiator implements connector.DbInstantiator with a sqlmock-backed *sqlx.DB
type fakeDbInstantiator struct {
	db *sqlx.DB
}

func (f *fakeDbInstantiator) GetConnection() *sqlx.DB { return f.db }
func (f *fakeDbInstantiator) CreateDatabase(_ string) {}
func (f *fakeDbInstantiator) DropDatabase(_ string)   {}
func (f *fakeDbInstantiator) GetDatabaseName() string { return "test" }
func (f *fakeDbInstantiator) GetDatabaseDriver() (database.Driver, error) {
	return nil, nil
}
func (f *fakeDbInstantiator) Connect() *sqlx.DB { return f.db }

func newTestContact() entity.Contact {
	return entity.Contact{
		Name:    "John",
		Message: "Hello",
		Email:   null.StringFrom("john@example.com"),
		DateAdd: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UUID:    "test-uuid",
	}
}

func TestMysqlContactRepository_Save(t *testing.T) {
	t.Run("Unit: test Save success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NilError(t, err)
		defer func() { _ = db.Close() }()

		mock.ExpectBegin()
		mock.ExpectPrepare(`INSERT INTO Contact`).
			ExpectExec().
			WithArgs("John", "Hello", "john@example.com", sqlmock.AnyArg(), "test-uuid").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		repo := repository.NewMysqlContactRepository(&fakeDbInstantiator{
			db: sqlx.NewDb(db, "sqlmock"),
		})

		err = repo.Save(context.Background(), newTestContact())

		assert.NilError(t, err)
		assert.NilError(t, mock.ExpectationsWereMet())
	})

	t.Run("Unit: test Save error begin transaction", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NilError(t, err)
		defer func() { _ = db.Close() }()

		mock.ExpectBegin().WillReturnError(sqlmock.ErrCancelled)

		repo := repository.NewMysqlContactRepository(&fakeDbInstantiator{
			db: sqlx.NewDb(db, "sqlmock"),
		})

		err = repo.Save(context.Background(), newTestContact())

		assert.ErrorContains(t, err, "error during begin transaction")
		assert.NilError(t, mock.ExpectationsWereMet())
	})

	t.Run("Unit: test Save error prepare statement", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NilError(t, err)
		defer func() { _ = db.Close() }()

		mock.ExpectBegin()
		mock.ExpectPrepare(`INSERT INTO Contact`).WillReturnError(sqlmock.ErrCancelled)

		repo := repository.NewMysqlContactRepository(&fakeDbInstantiator{
			db: sqlx.NewDb(db, "sqlmock"),
		})

		err = repo.Save(context.Background(), newTestContact())

		assert.ErrorContains(t, err, "error during prepare statement")
		assert.NilError(t, mock.ExpectationsWereMet())
	})

	t.Run("Unit: test Save error exec statement", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NilError(t, err)
		defer func() { _ = db.Close() }()

		mock.ExpectBegin()
		mock.ExpectPrepare(`INSERT INTO Contact`).
			ExpectExec().
			WithArgs("John", "Hello", "john@example.com", sqlmock.AnyArg(), "test-uuid").
			WillReturnError(sqlmock.ErrCancelled)

		repo := repository.NewMysqlContactRepository(&fakeDbInstantiator{
			db: sqlx.NewDb(db, "sqlmock"),
		})

		err = repo.Save(context.Background(), newTestContact())

		assert.ErrorContains(t, err, "error during exec statement")
		assert.NilError(t, mock.ExpectationsWereMet())
	})

	t.Run("Unit: test Save error commit transaction", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NilError(t, err)
		defer func() { _ = db.Close() }()

		mock.ExpectBegin()
		mock.ExpectPrepare(`INSERT INTO Contact`).
			ExpectExec().
			WithArgs("John", "Hello", "john@example.com", sqlmock.AnyArg(), "test-uuid").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit().WillReturnError(sqlmock.ErrCancelled)

		repo := repository.NewMysqlContactRepository(&fakeDbInstantiator{
			db: sqlx.NewDb(db, "sqlmock"),
		})

		err = repo.Save(context.Background(), newTestContact())

		assert.ErrorContains(t, err, "error during commit transaction")
		assert.NilError(t, mock.ExpectationsWereMet())
	})
}
