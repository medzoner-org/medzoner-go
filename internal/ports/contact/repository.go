//go:generate mockgen -destination=../../../test/mocks/contact_repository.go -package=mocks -source=./repository.go

package contact

import (
	"context"

	"github.com/Medzoner/medzoner-go/internal/domains"
)

type Repository interface {
	Save(ctx context.Context, contact domains.Contact) error
}
