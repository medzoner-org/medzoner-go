# Skill: Conventional Commits

## Description
Génère des messages de commit conformes à la spécification [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

## Format

```
<type>(<scope>): <description>

[body]

[footer(s)]
```

## Types

| Type       | Usage                                                        |
|------------|--------------------------------------------------------------|
| `feat`     | Ajout d'une nouvelle fonctionnalité                          |
| `fix`      | Correction de bug                                            |
| `docs`     | Modification de la documentation uniquement                  |
| `style`    | Changement de formatage (espaces, imports, etc.) sans impact |
| `refactor` | Refactoring du code sans ajout de feature ni correction      |
| `perf`     | Amélioration des performances                                |
| `test`     | Ajout ou modification de tests                               |
| `build`    | Changements liés au build ou aux dépendances                 |
| `ci`       | Changements de configuration CI/CD                           |
| `chore`    | Tâches de maintenance (mise à jour deps, nettoyage, etc.)    |
| `revert`   | Annulation d'un commit précédent                             |

## Scopes (spécifiques au projet)

| Scope        | Description                                                |
|--------------|------------------------------------------------------------|
| `cmd`        | Point d'entrée applicatif (`cmd/app`, `cmd/migrate`)       |
| `config`     | Configuration de l'application (`internal/config`)         |
| `domain`     | Entités et logique métier (`internal/domain`, `entity`)    |
| `handler`    | Handlers HTTP (`internal/ui/http/handler`)                 |
| `template`   | Templates HTML (`tmpl/`)                                   |
| `command`    | Command handlers CQRS (`internal/application/command`)     |
| `query`      | Query handlers CQRS (`internal/application/query`)         |
| `event`      | Event handlers (`internal/application/event`)              |
| `repository` | Repositories (`internal/domain/repository`)                |
| `migration`  | Migrations SQL (`migrations/`)                             |
| `db`         | Base de données (`internal/database`)                      |
| `wire`       | Injection de dépendances Wire (`internal/wire`)            |
| `docker`     | Dockerfiles et docker-compose (`infra/docker`, `infra/`)   |
| `k8s`        | Configuration Kubernetes (`infra/k8s`)                     |
| `ci`         | Pipelines CI/CD                                            |
| `deps`       | Dépendances Go (`go.mod`, `go.sum`)                        |
| `test`       | Tests et mocks (`test/`, `*_test.go`)                      |
| `mailer`     | Service d'envoi de mails                                   |
| `ui`         | Interface utilisateur (assets, CSS, images)                |

## Règles

1. **Description** : impératif, présent, minuscules, pas de point final, max 72 caractères
2. **Scope** : optionnel mais recommandé — utiliser les scopes ci-dessus
3. **Body** : optionnel — expliquer le *pourquoi* si nécessaire, séparé par une ligne vide
4. **Breaking changes** : ajouter `!` après le type/scope et un footer `BREAKING CHANGE:`
5. **Langue** : anglais pour les messages de commit
6. **Atomicité** : un commit = un changement logique

## Exemples

```
feat(handler): add contact form validation endpoint
```

```
fix(repository): handle nil pointer on empty contact query
```

```
refactor(command): extract contact creation logic into service
```

```
test(query): add unit tests for list techno query handler
```

```
build(deps): bump go.opentelemetry.io/otel to v1.43.0
```

```
ci: update GitHub Actions workflow for Go 1.25
```

```
docs: update README with local dev setup instructions
```

```
feat(domain)!: rename Contact entity fields

BREAKING CHANGE: Contact.Name split into FirstName and LastName
```

```
chore(docker): upgrade alpine base image to 3.23.3
```

```
fix(migration): correct column type in contacts table

The previous migration used VARCHAR(50) which truncated
long email addresses. Changed to VARCHAR(255).

Closes #42
```

## Instructions pour Claude

Quand on te demande de générer un commit :

1. Analyse les fichiers modifiés via `git diff --staged` (ou `git diff` si rien n'est stagé)
2. Détermine le **type** approprié selon la nature du changement
3. Identifie le **scope** en fonction des fichiers modifiés
4. Rédige une **description** concise et claire en anglais
5. Ajoute un **body** si le changement nécessite une explication
6. Si plusieurs changements logiques distincts sont détectés, propose plusieurs commits séparés
7. Exécute le commit avec `git commit -m "<message>"`

