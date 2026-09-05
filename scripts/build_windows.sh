#!/usr/bin/env bash
set -e

echo "Compiling not-jira for Windows x64 (windows/amd64)..."

export CGO_ENABLED=0
export GOOS=windows
export GOARCH=amd64

mkdir -p bin
go build -ldflags="-s -w" -trimpath -o bin/not-jira.exe cmd/bot/main.go

echo "Build successful! Output: bin/not-jira.exe"
ls -lh bin/not-jira.exe
