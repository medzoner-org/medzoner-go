package query

import (
	"context"
	"fmt"

	"github.com/Medzoner/gomedz/pkg/observability"
	"github.com/Medzoner/medzoner-go/internal/domain/repository"
	"go.opentelemetry.io/otel/attribute"
	otelTrace "go.opentelemetry.io/otel/trace"
)

// ListTechnoQueryHandler handles ListTechnoQuery and returns techno data.
type ListTechnoQueryHandler struct {
	TechnoRepository repository.TechnoRepository
}

// NewListTechnoQueryHandler creates a new ListTechnoQueryHandler.
func NewListTechnoQueryHandler(technoRepository repository.TechnoRepository) ListTechnoQueryHandler {
	return ListTechnoQueryHandler{
		TechnoRepository: technoRepository,
	}
}

// Handle handles ListTechnoQuery and returns a map of techno data.
func (l *ListTechnoQueryHandler) Handle(ctx context.Context, query ListTechnoQuery) (map[string]any, error) {
	ctx, iSpan := observability.StartSpan(ctx, "ListTechnoQueryHandler.Handle",
		otelTrace.WithAttributes(attribute.String("query.type", query.Type)),
	)
	defer iSpan.End()

	if query.Type != "stack" {
		return map[string]any{}, nil
	}

	result, err := l.TechnoRepository.FetchStack(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching stack: %w", err)
	}

	return result, nil
}
