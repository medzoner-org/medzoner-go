package http_utils

import (
	"context"
	"net/http"

	"github.com/Medzoner/gomedz/pkg/logger"
	otelTrace "go.opentelemetry.io/otel/trace"
)

// ResponseError writes an error response to the client and records the error in the span.
func ResponseError(w http.ResponseWriter, err error, code int, span otelTrace.Span) {
	span.RecordError(err)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	if _, werr := w.Write([]byte(err.Error())); werr != nil {
		logger.Error(context.TODO(), "failed to write error response", werr)
	}
}
