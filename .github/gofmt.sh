#!/bin/bash
set -ev

gofmt -w ./internal/ ./cmd/
test -z "$(gofmt -d -s ./internal/ ./cmd/ | tee /dev/stderr)"
