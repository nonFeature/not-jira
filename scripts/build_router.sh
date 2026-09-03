#!/usr/bin/env bash
set -e

echo "Compiling not-jira for OpenWrt (linux/arm64)..."

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=arm64

mkdir -p bin
go build -ldflags="-s -w" -trimpath -o bin/not-jira cmd/bot/main.go

echo "Build successful! Output: bin/not-jira"
ls -lh bin/not-jira
