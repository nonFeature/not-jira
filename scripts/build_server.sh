#!/usr/bin/env bash
set -e

echo "Compiling not-jira for Linux x86_64 (linux/amd64)..."

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

mkdir -p bin
go build -ldflags="-s -w" -trimpath -o bin/not-jira-amd64 cmd/bot/main.go

echo "Build successful! Output: bin/not-jira-amd64"
ls -lh bin/not-jira-amd64
