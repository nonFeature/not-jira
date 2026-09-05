# Скрипт сборки not-jira под Windows (windows/amd64)

Write-Host "Компиляция not-jira под Windows (windows/amd64)..." -ForegroundColor Cyan

$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

if (!(Test-Path -Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

go build -ldflags="-s -w" -trimpath -o bin/not-jira.exe cmd/bot/main.go

if ($LASTEXITCODE -eq 0) {
    $size = (Get-Item "bin/not-jira.exe").Length / 1MB
    Write-Host ("Сборка УСПЕШНА! Файл: bin/not-jira.exe ({0:N2} MB)" -f $size) -ForegroundColor Green
    Write-Host "Для запуска выполните: .\bin\not-jira.exe -config config.yaml" -ForegroundColor Yellow
} else {
    Write-Host "Ошибка компиляции!" -ForegroundColor Red
}
