package ports

import (
	"context"
)

type Repository[T any] interface {
	Save(ctx context.Context, obj T) error
}
