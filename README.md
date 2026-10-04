<p align="center">
  <img alt="golangci-lint logo" src="public/images/logo-single.png" height="150" />
  <h3 align="center">medzoner.com</h3>
  <p align="center">Source</p>
</p>

---
# Badges
[![Build Status](https://github.com/medzoner/medzoner-go/actions/workflows/github-actions.yml/badge.svg?branch=master)](https://github.com/medzoner/medzoner-go/actions/workflows/github-actions.yml/badge.svg?branch=master)
[![Coverage Status](https://coveralls.io/repos/github/Medzoner/medzoner-go/badge.svg?branch=master&service=github)](https://coveralls.io/github/Medzoner/medzoner-go?branch=master)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go report](https://goreportcard.com/badge/github.com/Medzoner/medzoner-go?service=github)](https://goreportcard.com/report/github.com/Medzoner/medzoner-go?service=github)

# Short Description
My website https://www.medzoner.com rewrite in golang from php (https://github.com/Medzoner/medzoner.com)

## Configuration

Configuration can be specified in .env or exported as environment variable:

| Name  | Value |
| ------------- | ------------- |
| ENV  | string  |
| DEBUG  | bool  |
| OPTIONS  | []string{} (option1,option2....)  |
| DATABASE_DSN  | string (ex: root:changeme@tcp(database:3366))  |
| DATABASE_NAME  | string  |
| DATABASE_DRIVER  | string  |
| MAILER_USER  | string  |
| MAILER_PASSWORD  | string  |

## Build

Requires Go 1.27.1+. The shared library is now
`github.com/medzoner-org/gomedz` (v1.0.14), not the old personal-account module.
For private module access, use a GitHub credential with read access and configure:

```sh
go env -w GOPRIVATE='github.com/Medzoner/*,github.com/medzoner-org/*'
```

Merge this with any existing private-module patterns rather than removing them.
The CI workflow configures both namespaces; installation of the public whyor
generator does not itself require private-module access.

```
go build -o bin/app ./cmd/app/main.go
go build -o bin/migrate ./cmd/migrate.go
```

## Run
```
go run ./cmd/app/main.go
go run ./cmd/migrate.go
```
![run](doc/run.png)

## Debug
```
go install github.com/go-delve/delve/cmd/dlv@master
dlv debug --headless --listen=:4000 --only-same-user=false --api-version=2 --accept-multiclient --log --log-output=rpc ./cmd/app/main.go
```
![run](doc/dlv.png)

## Test

### UnitTest
```
go test -v -cover -coverpkg=./... ./pkg/...
```
![run](doc/unit-test.png)

### BehaviorTest
Database (mysql) must be up and configured before run.
```
go test godog_test.go
```
![run](doc/behavior-test.png)

## Developed & Maintained by
[Mehdi Youb](https://github.com/Medzoner) 

## License 
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
