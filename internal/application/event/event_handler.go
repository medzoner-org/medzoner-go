package event

import "context"

// Handler is an interface that contains method Handle
type Handler interface {
	Publish(ctx context.Context, event Event) error
}
