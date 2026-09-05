# Скрипт сборки not-jira для роутера (OpenWrt / Linux aarch64)

Write-Host "Компиляция not-jira под OpenWrt (linux/arm64)..." -ForegroundColor Cyan

$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "arm64"

if (!(Test-Path -Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

go build -ldflags="-s -w" -trimpath -o bin/not-jira cmd/bot/main.go

if ($LASTEXITCODE -eq 0) {
    $size = (Get-Item "bin/not-jira").Length / 1MB
    Write-Host ("Сборка УСПЕШНА! Файл: bin/not-jira ({0:N2} MB)" -f $size) -ForegroundColor Green
    Write-Host "Команда для отправки на роутер:" -ForegroundColor Yellow
    Write-Host "  scp bin/not-jira root@192.168.1.1:/usr/bin/not-jira" -ForegroundColor White
} else {
    Write-Host "Ошибка компиляции!" -ForegroundColor Red
}
