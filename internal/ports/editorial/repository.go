//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination=../../../test/mocks/editorial_repository.go -package=mocks -mock_names Repository=EditorialRepository -source=repository.go

package editorial

import (
	"context"

	"github.com/Medzoner/medzoner-go/internal/domains/editorial"
)

// Repository provides the editorial source; publication rules belong to queries.
type Repository interface {
	All(context.Context) ([]editorial.Article, error)
}
