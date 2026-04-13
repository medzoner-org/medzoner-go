package repository

import (
	"context"
	"fmt"
	"io"
	"reflect"

	sq "github.com/Masterminds/squirrel"
	"github.com/Medzoner/gomedz/pkg/connector"
	"github.com/Medzoner/gomedz/pkg/logger"
	"github.com/Medzoner/gomedz/pkg/observability"
	"github.com/Medzoner/medzoner-go/internal/entity"
	otelTrace "go.opentelemetry.io/otel/trace"
)

// MysqlContactRepository MysqlContactRepository
type MysqlContactRepository struct {
	DbInstance connector.DbInstantiator
}

// NewMysqlContactRepository is a function that returns a new MysqlContactRepository
func NewMysqlContactRepository(dbInstance connector.DbInstantiator) *MysqlContactRepository {
	return &MysqlContactRepository{
		DbInstance: dbInstance,
	}
}

// Save is a function that saves a contact
func (m *MysqlContactRepository) Save(ctx context.Context, contact entity.Contact) error {
	_, iSpan := observability.StartSpan(ctx, "MysqlContactRepository.Save")
	defer iSpan.End()

	conn, err := m.DbInstance.GetConnection().Begin()
	if err != nil {
		return errorResponse("error during begin transaction", iSpan, err)
	}

	query, args, err := sq.Insert("Contact").
		Columns("name", "message", "email", "date_add", "uuid").
		Values(contact.Name, contact.Message, contact.EmailValue(), contact.DateAdd, contact.UUID).
		ToSql()
	if err != nil {
		return errorResponse("error during build query", iSpan, err)
	}

	stmt, err := conn.Prepare(query)
	defer closeStmt(ctx, stmt, iSpan)
	if err != nil {
		return errorResponse("error during prepare statement", iSpan, err)
	}

	_, err = stmt.Exec(args...)
	if err != nil {
		return errorResponse("error during exec statement", iSpan, err)
	}

	if err = conn.Commit(); err != nil {
		return errorResponse("error during commit transaction", iSpan, err)
	}

	return nil
}

func errorResponse(msg string, iSpan otelTrace.Span, err error) error {
	iSpan.RecordError(err)
	return fmt.Errorf("%s: %w", msg, err)
}

func closeStmt(ctx context.Context, stmt io.Closer, span otelTrace.Span) {
	if stmt == nil || reflect.ValueOf(stmt).IsNil() {
		return
	}
	if err := stmt.Close(); err != nil {
		span.RecordError(err)
		logger.Error(ctx, "stmt close error.", err)
	}
}
