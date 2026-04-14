.DEFAULT_GOAL := help
.PHONY: help setup check-tools githooks test_all build start migrate wire generate \
        lint lint-fix govet gofmt staticcheck gosec gocyclo ineffassign run-qa \
        docker-up docker-down skaffold-run coverage trace trivy k6

ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# ─── Help ────────────────────────────────────────────────────────────────────
help: ## Affiche cette aide
	@echo ""
	@echo "  medzoner-go — Commandes disponibles"
	@echo "  ════════════════════════════════════"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sed 's/Makefile://' | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
	@echo ""

# ─── Setup ───────────────────────────────────────────────────────────────────
setup: check-tools githooks ## Setup complet du projet (outils, hooks, env)
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "📄 .env créé depuis .env.example — pensez à le configurer"; \
	else \
		echo "📄 .env existe déjà"; \
	fi
	@go mod download
	@echo "✅ Setup terminé"

check-tools: ## Vérifie les outils de dev requis
	@chmod +x scripts/check-tools.sh
	@./scripts/check-tools.sh

githooks: ## Configure les hooks Git
	@git config core.hooksPath .githooks
	@chmod +x .githooks/*
	@echo "🪝 Git hooks activés (.githooks/)"

# ─── Build & Run ─────────────────────────────────────────────────────────────
build: ## Build des binaires (app + migrate)
	CGO_ENABLED=0 go build -o ./bin/app ./cmd/app/main.go
	CGO_ENABLED=0 go build -o ./bin/migrate ./cmd/migrate/migrate.go

start: ## Démarre le serveur HTTP
	go run ./cmd/app/main.go

migrate: ## Lance les migrations DB (up)
	go run ./cmd/migrate/migrate.go

# ─── Code Generation ─────────────────────────────────────────────────────────
wire: ## Régénère l'injection de dépendances (Wire)
	wire gen ./internal/wire/

generate: ## Régénère les mocks (mockgen)
	go generate ./...

# ─── Tests ───────────────────────────────────────────────────────────────────
test_all: ## Tests unitaires avec couverture
	go test -v -count=1 -cover -coverpkg=./... -covermode=count -coverprofile=cov.out ./...
	go tool cover -func=cov.out

coverage: test_all ## Tests + ouvre le rapport HTML de couverture
	go tool cover -html=cov.out -o coverage.html
	@echo "📊 Rapport: coverage.html"

# ─── Qualité de code ─────────────────────────────────────────────────────────
lint: ## Lint via golangci-lint
	golangci-lint -v --config .golangci/.golangci.yml --issues-exit-code 1 run ./...

lint-fix: ## Lint avec auto-fix
	golangci-lint -v --config .golangci/.golangci.yml --issues-exit-code=1 run --fix ./...

govet: ## Go vet
	go vet ./...

gofmt: ## Vérifie le formatage Go
	@gofmt -w ./internal/ ./cmd/
	@test -z "$$(gofmt -d -s ./internal/ ./cmd/ | tee /dev/stderr)"

staticcheck: ## Analyse statique (staticcheck)
	staticcheck --debug.version
	staticcheck ./...

gosec: ## Audit de sécurité (gosec)
	gosec ./...

gocyclo: ## Complexité cyclomatique
	gocyclo -ignore "_test|Godeps|var|vendor/" .

ineffassign: ## Détecte les assignations inutiles
	ineffassign ./...

run-qa: govet gofmt lint staticcheck gosec gocyclo ## QA complète (tous les checks)
	@echo "✅ QA passed"

# ─── Docker ──────────────────────────────────────────────────────────────────
docker-up: ## Démarre les services Docker (MariaDB, Mailhog)
	docker compose -f infra/docker/local/docker-compose.yml up -d

docker-down: ## Arrête les services Docker
	docker compose -f infra/docker/local/docker-compose.yml down

# ─── Infra ───────────────────────────────────────────────────────────────────
skaffold-run: ## Déploie sur K8s via Skaffold
	skaffold dev --port-forward --platform=linux/arm64,linux/amd64 --insecure-registry=registry.medzoner.lan:5000

# ─── Divers ──────────────────────────────────────────────────────────────────
trace: ## Ouvre une trace Go
	go tool trace trace.out

trivy: ## Scan de sécurité container (Trivy)
	trivy image

k6: ## Test de charge K6
	docker run --rm -i grafana/k6 run --vus 10 --duration 30s - <test/k6/test.js
