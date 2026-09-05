@echo off
chcp 65001 > nul
echo Компиляция not-jira под Windows (windows/amd64)...

set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64

if not exist "bin" mkdir "bin"

go build -ldflags="-s -w" -trimpath -o bin\not-jira.exe cmd\bot\main.go

if %ERRORLEVEL% equ 0 (
    echo Сборка УСПЕШНА! Файл: bin\not-jira.exe
    echo Для запуска выполните: bin\not-jira.exe -config config.yaml
) else (
    echo Ошибка компиляции!
)
pause
