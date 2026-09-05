# Скрипт сборки not-jira под обычный Linux сервер / хостинг (linux/amd64)

Write-Host "Компиляция not-jira под Linux x86_64 (linux/amd64)..." -ForegroundColor Cyan

$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"

if (!(Test-Path -Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

go build -ldflags="-s -w" -trimpath -o bin/not-jira-amd64 cmd/bot/main.go

if ($LASTEXITCODE -eq 0) {
    $size = (Get-Item "bin/not-jira-amd64").Length / 1MB
    Write-Host ("Сборка УСПЕШНА! Файл: bin/not-jira-amd64 ({0:N2} MB)" -f $size) -ForegroundColor Green
    Write-Host "Этот бинарник подходит для большинства Linux-серверов и хостингов." -ForegroundColor Yellow
} else {
    Write-Host "Ошибка компиляции!" -ForegroundColor Red
}
