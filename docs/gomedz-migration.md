# gomedz migration and OpenTelemetry security correction

The application moves from `github.com/Medzoner/gomedz v0.14.11` to
`github.com/medzoner-org/gomedz v1.0.14`. All source imports and generated wiring
use the organization module path. No local `replace`, workspace override or
module-cache edit is required.

## Compatibility changes

- Validation's `Validater` interface becomes `Validator`; the concrete
  `ValidatorAdapter` becomes `Adapter`.
- `validation.New` now accepts variadic options. The composition root uses a
  non-variadic `newValidator` wrapper so existing whyor wiring remains explicit.
- Existing HTTP, application and repository behavior is covered by the unit
  and Godog scenarios; the migration does not redesign handlers or lifecycle.
- CI private-module patterns include both the legacy and organization namespaces.
- The v1.0.14 BDD helper caches HTTP response bodies before closing the network
  stream, preserving the content assertions in all existing scenarios.

## Security scope

The corrected library upgrades OpenTelemetry stable modules/exporters to
1.45.0 and logging modules to 0.21.0, addressing:

- GO-2026-6615: log batch processor busy spin.
- GO-2026-6508: log exporter environment TLS configuration.
- GO-2026-6505: exporter endpoint information in logs.

The logger is adapted to the new attribute value API. Simply forcing the new
logging module onto gomedz v0.14.11 does not compile.

This migration is not a blanket claim of zero vulnerability findings. A symbol
scan distinguishes called vulnerable symbols from imported packages and module
presence. Residual grpc or unmaintained OpenPGP findings must be reviewed rather
than hidden. grpc is not moved to a development revision in this migration.

After migration, the local symbol scan reports **0 called vulnerabilities**.
Two findings remain without a detected call path:

- GO-2026-6443 in the imported grpc package; its reported fix is a development
  revision, so grpc stays at the stable version already used by this application.
- GO-2026-5932 concerns the unmaintained OpenPGP package within the x/crypto
  module; its presence is reported, not treated as proof the application uses it.

compress is updated to 1.18.7 and chi to 5.3.0 to remove their module findings.

## Existing race-detector limitation

`go test -race ./...` detects an internal race between Migration.Buffer and
Migrate.runMigrations in golang-migrate 4.19.1, exercised by pkg/database tests.
It is also reproduced on the pre-migration master commit `4bf5944` using
`go test -race ./pkg/database -count=3`; the dependency version is unchanged.
This migration does not suppress that finding or claim a clean full race run.

## Verify

```sh
go build ./...
go vet ./...
go test ./...
whyor check ./internal/wire
govulncheck ./...
```
