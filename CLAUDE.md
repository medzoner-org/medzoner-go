# CLAUDE.md — medzoner-go

## Projet

Site web personnel (portfolio/contact) en Go, architecture hexagonale (DDD), avec formulaire de contact, affichage de technos, et envoi de mails. Déployé sur Kubernetes (k3s) via Skaffold.

## Stack technique

- **Go 1.26+** — binaires dans `cmd/app` (serveur HTTP) et `cmd/migrate` (migrations DB)
- **Framework HTTP** : `github.com/Medzoner/gomedz` (wrapper Gin)
- **DI** : Google Wire (`internal/wire/`)
- **DB** : MySQL/MariaDB via `jmoiron/sqlx` + `go-sql-driver/mysql`
- **Migrations** : `golang-migrate/migrate/v4` (source `file://`, driver `mysql`)
- **Observabilité** : OpenTelemetry (traces, métriques, logs)
- **Mail** : SMTP via `gomedz/pkg/notifier`
- **Config** : variables d'environnement via `caarlos0/env/v11`
- **Templates** : HTML Go standard (`tmpl/`)
- **Captcha** : reCAPTCHA via `gomedz/pkg/captcha`
- **Tests BDD** : Cucumber/Godog (`test/features/`)
- **Mocks** : `go.uber.org/mock` (mockgen) + `DATA-DOG/go-sqlmock`
- **Assertions** : `gotest.tools/assert`
- **Lint** : golangci-lint (config `.golangci/.golangci.yml`)
- **Infra** : Docker, Docker Compose, Kubernetes, Skaffold

## Architecture

```
cmd/
  app/main.go              → point d'entrée serveur HTTP
  migrate/migrate.go       → point d'entrée migrations

internal/
  config/                  → configuration (env vars)
  database/                → migrations DB (golang-migrate)
  entity/                  → entités (Contact)
  domain/
    customtype/            → types personnalisés (NullString)
    repository/            → interfaces + implémentations (MySQL, JSON)
  application/
    command/               → CQRS commands (CreateContactCommandHandler)
    query/                 → CQRS queries (ListTechnoQueryHandler)
    event/                 → domain events (ContactCreatedEventHandler)
    service/mailer/        → interface Mailer
  ui/http/
    handler/               → HTTP handlers (IndexHandler, NotFoundHandler)
    http_utils/            → utilitaires HTTP (ResponseError)
    templater/             → rendering HTML
  wire/                    → injection de dépendances (Wire)
  resources/data/          → données statiques (JSON technos)

test/
  mocks.go                 → agrégateur de mocks (struct Mocks)
  mocks/                   → mocks générés (mockgen)
  features/
    bootstrap/             → contexte Godog
    test/                  → fichiers .feature (Gherkin)

migrations/                → fichiers SQL (up/down)
tmpl/                      → templates HTML
public/                    → assets statiques (CSS, images)
```

## Commandes

```bash
# Setup initial (outils, hooks, env)
make setup

# Vérifier les outils installés
make check-tools

# Démarrer le serveur
make start

# Build des binaires
make build

# Lancer les migrations (up par défaut)
make migrate
go run ./cmd/migrate/migrate.go down

# Tests unitaires avec couverture
make test_all

# Tests + rapport HTML de couverture
make coverage

# Tests BDD (Godog) — nécessite le serveur
GODOG_INTEGRATION=1 go test -v ./...

# Lint
make lint
make lint-fix

# Analyse statique
make staticcheck
make gosec
make gocyclo

# QA complète
make run-qa

# Wire (régénérer l'injection)
make wire

# Régénérer les mocks
make generate

# Docker local
make docker-up
make docker-down

# Skaffold (K8s)
make skaffold-run

# Aide (liste toutes les commandes)
make help
```

## Conventions de code

### Nommage
- Packages : snake_case simple (un mot si possible)
- Fichiers : `snake_case.go`, tests : `snake_case_test.go`
- Structs/interfaces : PascalCase
- Constructeurs : `New<Type>(deps) <Type>` ou `*<Type>`
- Tests : `Test<Struct>_<Method>(t *testing.T)` avec sous-tests `t.Run("Unit: test <description>", ...)`

### Tests
- Package de test : `<package>_test` (black-box testing)
- Assertions : **`gotest.tools/assert`** — ne pas utiliser `testify`
  - `assert.NilError(t, err)`
  - `assert.Equal(t, got, want)`
  - `assert.ErrorContains(t, err, "substring")`
  - `assert.Error(t, err, "exact message")`
  - `assert.Assert(t, condition)`
- Mocks : créés via `mockgen` dans `test/mocks/`, agrégés dans `test/mocks.go` (`mocks.New(t)`)
- L'init des tests nécessitant l'observabilité doit appeler `observability.NewTelemetry(...)` dans `func init()`
- Utiliser `t.TempDir()` pour les fichiers temporaires
- Utiliser `t.Helper()` dans les fonctions helper de test
- Utiliser `t.Setenv()` pour les variables d'environnement de test

### Patterns
- **CQRS** : Commands (`Handle(ctx, Command) error`) et Queries (`Handle(ctx, Query) (result, error)`)
- **Events** : `Event` interface → `Publish(ctx, Event) error`
- **Repository** : interfaces dans `domain/repository/`, implémentations dans le même package
- **Wire** : constructeurs sans logique, tous les wirings dans `internal/wire/wire.go`
- **Observabilité** : `observability.StartSpan(ctx, "name")` dans chaque méthode publique, `defer span.End()`
- **Erreurs** : `fmt.Errorf("description: %w", err)` — toujours wrapper avec contexte

### Gestion des erreurs
- Toujours wrapper les erreurs avec `fmt.Errorf("context: %w", err)`
- Les handlers HTTP utilisent `http_utils.ResponseError(w, err, status, span)`
- Les repository enregistrent l'erreur dans le span : `span.RecordError(err)`

### Mocks (go:generate)
- Directive en tête de fichier d'interface :
  ```go
  //go:generate mockgen -destination=../../../test/mocks/<name>.go -package=mocks -source=./<interface_file>.go
  ```

## Variables d'environnement

| Préfixe | Usage |
|---|---|
| `TELEMETRY_` | OpenTelemetry |
| `ENGINE_` | Gin engine |
| `LOGGER_` | Logging (zerolog) |
| `AUTH_` | JWT / SSO |
| `SERVER_` | HTTP server |
| `MAILER_` | SMTP config |
| `DATABASE_` | MySQL DSN, nom, driver |
| `RECAPTCHA_` | reCAPTCHA keys |
| `ROOT_PATH` | Chemin racine du projet |

## Base de données

- **MySQL/MariaDB** — table `Contact` (id, uuid, name, email, message, date_add)
- Migrations dans `migrations/` — format : `<timestamp>_<name>.(up|down).sql`
- Driver : `golang-migrate/migrate/v4` avec source `file://` et database `mysql`

## CI / Qualité

### GitHub Actions (`.github/workflows/ci.yml`)
- **Déclenchement** : push sur `master`, `main`, `develop`, `feat/**`, `fix/**`, `refactor/**`, `release/**` + PRs
- **Jobs** :
  - `lint` — golangci-lint avec config `.golangci/.golangci.yml`
  - `test` — tests unitaires avec couverture (MariaDB en service container)
  - `build` — compilation des binaires (après lint + test)
  - `security` — audit gosec
- **Concurrency** : annule les runs précédents sur la même branche

### Local
- Pre-commit hooks : `gofmt` + `go vet` + `go test` (via `.githooks/pre-commit`)
- Commit-msg hook : validation Conventional Commits (via `.githooks/commit-msg`)
- Activation : `make githooks` ou `make setup`
- Linters actifs : errcheck, gosimple, govet, staticcheck, gosec, cyclop, wrapcheck, spancheck, etc.
- Les fichiers `_test.go` sont exemptés de : cyclop, gosec, dupl, funlen, govet, prealloc, gocritic
- Couverture : `go test -cover -coverpkg=./... -covermode=count`

## Points d'attention

- **Blank imports requis** dans `db_migration.go` : `_ "github.com/golang-migrate/migrate/v4/source/file"` et `_ "github.com/go-sql-driver/mysql"`
- Wire génère `wire_gen.go` — ne **jamais** éditer manuellement, relancer `wire gen`
- Les mocks sont auto-générés — relancer `go generate ./...` après modification d'interfaces
- Le serveur de test Godog utilise des mocks Wire injectés (`InitServerTest`)

