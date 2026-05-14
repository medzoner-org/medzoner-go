# AGENTS.md — medzoner-go

## Architecture

Hexagonal (DDD) Go portfolio site with CQRS. Request flow: `HTTP handler → Command/Query → Repository/Event → DB/Mail`.

- **DI**: Google Wire (`internal/wire/wire.go`) — never edit `wire_gen.go`, run `make wire`
- **HTTP framework**: `github.com/Medzoner/gomedz` (wraps Fiber), handlers implement `http.Controller` (see `Register()` method pattern in `internal/ui/http/handler/index_handler.go`)
- **Config**: env vars parsed via `caarlos0/env/v11` into `internal/config/config.go`, prefixed (`TELEMETRY_`, `DATABASE_`, `MAILER_`, etc.)

## Key Patterns

### CQRS Commands & Events
Commands: `Handle(ctx, Command) error` — see `internal/application/command/create_contact_command_handler.go`
Events: `Publish(ctx, Event) error` — see `internal/application/event/contact_created_event_handler.go`
Each public method must start with `ctx, span := observability.StartSpan(ctx, "Type.Method")` + `defer span.End()`.

### SQL — No Raw Queries
Always use Squirrel (`sq.Insert`, `sq.Select`). See `internal/domain/repository/contact_mysql.go`.

### Error Handling
Always wrap: `fmt.Errorf("context: %w", err)`. In repos, also `span.RecordError(err)`. In HTTP handlers, use `http_utils.ResponseError(w, err, status, span)`.

### Interfaces & Mocks
Interfaces live in `internal/domain/` with `//go:generate mockgen` directives at top. Mocks output to `test/mocks/`. Aggregated in `test/mocks.go` (`mocks.New(t)`). Regenerate: `make generate`.

## Testing Conventions

- Package: `<pkg>_test` (black-box). Test names: `t.Run("Unit: test <description>", ...)`
- Assertions: **`gotest.tools/assert`** only (no testify). Use `assert.NilError`, `assert.Equal`, `assert.Error`, `assert.ErrorContains`
- Tests needing tracing require `func init()` with `observability.NewTelemetry(...)` — see `create_contact_command_handler_test.go`
- Mock setup: `mocked := mocks.New(t)` then `mocked.Mailer.EXPECT().Send(...)`

## Commands

```bash
make test_all        # unit tests + coverage
make lint            # golangci-lint (config: .golangci/.golangci.yml)
make wire            # regenerate DI
make generate        # regenerate mocks
make build           # build binaries to bin/
make docker-up       # start local MariaDB + Mailhog
make run-qa          # full QA (vet, fmt, lint, staticcheck, gosec, gocyclo)
```

BDD tests: `GODOG_INTEGRATION=1 go test -v ./...` (requires running DB).

## File Naming

- Files: `snake_case.go`, tests: `snake_case_test.go`
- Constructors: `New<Type>(deps) *<Type>` or value type
- Wire sets grouped by layer in `internal/wire/wire.go`: `CommonWiring`, `DbWiring`, `RepositoryWiring`, `UsecaseWiring`, `HandlerWiring`, `ServerWiring`

## Gotchas

- `internal/database/db_migration.go` requires blank imports: `_ "github.com/golang-migrate/migrate/v4/source/file"` and `_ "github.com/go-sql-driver/mysql"`
- Lint config at `.golangci/.golangci.yml` (not root) — test files are exempt from cyclop, gosec, dupl, funlen
- Migrations in `migrations/` use format `<timestamp>_<name>.(up|down).sql`
- Git hooks: Conventional Commits enforced via `.githooks/commit-msg`

